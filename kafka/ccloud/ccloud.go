package ccloud

/**
 * Copyright 2020 Confluent Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

var configFile *string
var topic *string

const (
	METADATA_BROKER_LIST            = "metadata.broker.list"
	BOOTSTRAP_SERVERS               = "bootstrap.servers"
	SASL_MECHANISMS                 = "sasl.mechanisms"
	SECURITY_PROTOCOL               = "security.protocol"
	SASL_USERNAME                   = "sasl.username"
	SASL_PASSWORD                   = "sasl.password"
	GROUP_ID                        = "group.id"
	GO_EVENTS_CHANNEL_ENABLE        = "go.events.channel.enable"
	GO_APPLICATION_REBALANCE_ENABLE = "go.application.rebalance.enable"
	NUM_PARTITIONS                  = "num.partitions"
	REPLICATION_FACTOR              = "replication.factor"
)

// ParseArgs parses the command line arguments and
// returns the config file and topic on success, or exits on error
func ParseArgs() (*string, *string) {
	if flag.Lookup("f") == nil {
		configFile = flag.String("f", "", "Path to Confluent Cloud configuration file")
	}
	if flag.Lookup("t") == nil {
		topic = flag.String("t", "", "Topic name")
	}

	flag.Parse()
	if *configFile == "" || *topic == "" {
		flag.Usage()
		os.Exit(2) // the same exit code flag.Parse uses
	}

	return configFile, topic

}

// ReadCCloudConfig reads the file specified by configFile and
// creates a map of key-value pairs that correspond to each
// line of the file. ReadCCloudConfig returns the map on success,
// or exits on error
func ReadCCloudConfig(configFile string) map[string]string {

	m := make(map[string]string)

	file, err := os.Open(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open file: %s", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && len(line) != 0 {
			kv := strings.Split(line, "=")
			parameter := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			m[parameter] = value
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Failed to read file: %s", err)
		os.Exit(1)
	}

	return m

}
