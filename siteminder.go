package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Config Object for Siteminder
type Config struct {
	InsightsURL       string            `yaml:"insights_url"`
	InsightsKey       string            `yaml:"insights_key"`
	Port              string            `yaml:"insights_key"`
	Interval          string            `yaml:"interval"`
	CustomAttributes  map[string]string `yaml:"custom_attributes"`
	Debug             bool              `yaml:"debug"`
	MaxBufferSize     int               `yaml:"max_buffer_size"`
	MaxRequestRetries int               `yaml:"max_request_retries"`
	ProxyURL          string            `yaml:"proxy_url"`
}

// Siteminder is the main Object where data is collected
type Siteminder struct {
	Config     Config
	Queue      chan Event
	HTTPClient *http.Client
}

// Metric is the Object containing the metric data
type Metric struct {
	MetricName  string `xml:"name,attr"`
	MetricValue string `xml:"value,attr"`
	MetricType  string `xml:"type,attr"`
}

// Event contains the data to send to NR
type Event struct {
	EventData       map[string]string
	EventSize       int
	NumberOfRetries int
}

var logFile *os.File
var logger *log.Logger

const (
	agentVersion      = "1.0.0"
	eventType         = "siteminderSample"
	configFileName    = "siteminder.yml"
	logFileName       = "siteminder.log"
	maxRetriesDefault = 5
	maxBufferDefault  = 100
	intervalDefault   = "30s"
)

func main() {

	logInit()
	defer logFile.Close()

	config := getConfig()
	siteminder := &Siteminder{
		Config:     config,
		Queue:      make(chan Event),
		HTTPClient: getHTTPClient(config.ProxyURL),
	}

	go siteminder.readQueue()
	siteminder.listener()
}

func (siteminder *Siteminder) listener() {
	l, err := net.Listen("tcp", siteminder.Config.Port)
	if err != nil {
		logger.Fatal("Error occured when trying to listen to the port:", siteminder.Config.Port, "==>", err)
	}
	defer l.Close()

	c, err := l.Accept()
	if err != nil {
		logger.Fatal("Error occured when trying to accept a connection:", err)
	}

	for {
		scanner := bufio.NewScanner(c)
		for scanner.Scan() {
			line := scanner.Text()
			debug("Reading one line:", line, siteminder.Config.Debug)
			siteminder.extractMetric(line)
		}
	}
}

func (siteminder *Siteminder) extractMetric(line string) {
	re := regexp.MustCompile(`<metric .*?.\/>`)
	if re.MatchString(line) {
		metric := Metric{}
		submatchall := re.FindAllString(line, -1)
		for _, element := range submatchall {
			err := xml.Unmarshal([]byte(element), &metric)
			if err != nil {
				logger.Println("Error parsing the metric xml data in one line:", err)
			} else {
				debug("Extracting metrics:", metric, siteminder.Config.Debug)
				siteminder.buildEvent(metric)
			}
		}
	} else {
		debug("No match for this pattern:", `<metric .*?.\/>`, siteminder.Config.Debug)
	}
}

func (siteminder *Siteminder) buildEvent(metricObj Metric) {
	metricEvent := map[string]string{}
	metricEvent["eventType"] = eventType
	metricEvent["agentVersion"] = agentVersion
	metricEvent["interval"] = siteminder.Config.Interval
	metricEvent["port"] = siteminder.Config.Port
	for k, v := range siteminder.Config.CustomAttributes {
		metricEvent[k] = v
	}
	metricEvent["metricType"] = metricObj.MetricType
	metricEvent["metricName"] = metricObj.MetricName
	metricEvent["metricValue"] = metricObj.MetricValue

	metricEventJSON, err := json.Marshal(metricEvent)
	if err != nil {
		logger.Println("Error: failed to Marshal 1 event:", err)
	} else {
		debug("Adding event to the queue:", metricEvent, siteminder.Config.Debug)
		siteminder.Queue <- Event{
			EventData:       metricEvent,
			EventSize:       len(string(metricEventJSON)),
			NumberOfRetries: 0,
		}
	}
}

func (siteminder *Siteminder) readQueue() {
	buffer := make([]Event, 0)
	duration, err := time.ParseDuration(siteminder.Config.Interval)
	if err != nil {
		logger.Println("Failed to parse interval, using default instead")
		duration, _ = time.ParseDuration(intervalDefault)
	}
	timeout := time.NewTimer(duration)
	byteSize := 0

	for {
		select {
		case msg := <-siteminder.Queue:
			if byteSize >= siteminder.Config.MaxBufferSize {
				debug("Flushing buffer because of size:", byteSize, siteminder.Config.Debug)
				timeout.Stop()
				siteminder.flushBuffer(buffer)
				timeout.Reset(duration)
				buffer = make([]Event, 0)
				byteSize = 0
			}
			buffer = append(buffer, msg)
			byteSize += msg.EventSize

		case <-timeout.C:
			if len(buffer) > 0 {
				debug("Flushing buffer because the interval limit is reached:", duration, siteminder.Config.Debug)
				siteminder.flushBuffer(buffer)
				timeout.Reset(duration)
				buffer = make([]Event, 0)
				byteSize = 0
			} else {
				timeout.Reset(duration)
			}
		}
	}
}

func (siteminder *Siteminder) flushBuffer(events []Event) {
	var data bytes.Buffer
	eventsArray := make([]map[string]string, 0)
	for _, event := range events {
		eventsArray = append(eventsArray, event.EventData)
	}
	if error := json.NewEncoder(&data).Encode(eventsArray); error != nil {
		logger.Println("JSON Encoding Error:", error)
		return
	}
	debug("Sending payload:", &data, siteminder.Config.Debug)
	_, err := url.Parse(siteminder.Config.InsightsURL)
	if err != nil {
		logger.Fatal("Parsing Insights URL Error:", err)
	}
	req, err := http.NewRequest("POST", siteminder.Config.InsightsURL, &data)
	if err != nil {
		logger.Println("http: error on http.NewRequest:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Insert-Key", siteminder.Config.InsightsKey)

	resp, err := siteminder.HTTPClient.Do(req)
	if err != nil {
		if _, ok := err.(net.Error); ok {
			go siteminder.retry(events)
		} else {
			logger.Println("HTTP Client Post Request Error:", err)
		}
		return
	}
	if resp != nil {
		if resp.StatusCode != http.StatusOK {
			logger.Println("Received Status Code:", resp.StatusCode, "While Sending Message")
		}
		defer resp.Body.Close()
	}
}

func (siteminder *Siteminder) retry(buffer []Event) {
	for _, event := range buffer {
		if event.NumberOfRetries < siteminder.Config.MaxRequestRetries {
			debug("Retrying to send event again because of a network failure:", event, siteminder.Config.Debug)
			event.NumberOfRetries++
			siteminder.Queue <- event
		} else {
			debug("Maximum retries limit reached:", event, siteminder.Config.Debug)
		}
	}
}

func getConfig() Config {
	data, err := ioutil.ReadFile(configFileName)
	if err != nil {
		logger.Fatal("Error reading the config file:", err)
	}
	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		logger.Fatal("Error unmarshalling the config file:", err)
	}
	checkConfig(&config)
	debug("Configuration attributes:", config, config.Debug)
	return config
}

func checkConfig(config *Config) {
	// Overwrite if environment variable is provided
	if url := os.Getenv("NR_INSIGHTS_URL"); url != "" {
		config.InsightsURL = url
	} else if config.InsightsURL == "" {
		logger.Fatal("Required attribute: Insights URL is mssing in your configuration file")
	}
	if key := os.Getenv("NR_INSIGHTS_KEY"); key != "" {
		config.InsightsKey = key
	} else if config.InsightsKey == "" {
		logger.Fatal("Required attribute: Insights Key is mssing in your configuration file")
	}
	if port := os.Getenv("NR_PORT"); port != "" {
		config.Port = port
	} else if config.Port == "" {
		logger.Fatal("Required attribute: Port number is mssing in your configuration file")
	}
	if interval := os.Getenv("NR_INTERVAL"); interval != "" {
		config.Interval = interval
	} else if config.Interval == "" {
		config.Interval = intervalDefault
	}
	if debug := os.Getenv("NR_DEBUG"); debug != "" {
		b, err := strconv.ParseBool(debug)
		if err != nil {
			logger.Println("Cannot parse DEBUG to boolean, using default instead")
		}
		config.Debug = b
	}
	if maxBuffersize := os.Getenv("NR_MAX_BUFFER_SIZE"); maxBuffersize != "" {
		bufferSize, err := strconv.Atoi(maxBuffersize)
		if err != nil {
			logger.Println("Cannot parse MAX_BUFFER_SIZE to int, using default instead")
		}
		config.MaxBufferSize = bufferSize * 1024 // KB
	} else if config.MaxBufferSize == 0 {
		config.MaxBufferSize = maxBufferDefault * 1024 // KB
	}
	if maxRequestRetries := os.Getenv("NR_MAX_REQUEST_RETRIES"); maxRequestRetries != "" {
		maxRetries, err := strconv.Atoi(maxRequestRetries)
		if err != nil {
			logger.Println("Cannot parse MAX_REQUEST_RETRIES to int,using default instead")
		}
		config.MaxRequestRetries = maxRetries
	} else if config.MaxRequestRetries == 0 {
		config.MaxRequestRetries = maxRetriesDefault
	}
	if proxy := os.Getenv("NR_PROXY_URL"); proxy != "" {
		config.InsightsURL = proxy
	}
	if customAttrs := os.Getenv("NR_CUSTOM_ATTRS"); customAttrs != "" {
		config.CustomAttributes = getCustomAttrs(customAttrs)
	}
}

func getCustomAttrs(tags string) map[string]string {
	m := make(map[string]string)
	arrayTags := strings.Split(tags, ";")
	for _, tag := range arrayTags {
		pairs := strings.Split(tag, ":")
		if len(pairs) > 1 {
			m[pairs[0]] = pairs[1]
		}
	}
	return m
}

func getHTTPClient(proxyURLValue string) *http.Client {
	transport := &http.Transport{}
	if proxyURLValue != "" {
		proxyURL, err := url.Parse(proxyURLValue)
		if err != nil {
			logger.Fatal("Parsing Proxy URL Error:", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &http.Client{Transport: transport, Timeout: time.Second * 60} // Timeout after 60s
}

func logInit() {
	logFile, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	logger = log.New(logFile, "siteminderLog: ", log.LstdFlags)
}

func debug(msg string, content interface{}, debugIsOn bool) {
	if debugIsOn {
		logger.Println("DEBUG mode on:", msg, content)
	}
}
