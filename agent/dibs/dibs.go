package dibs

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

const (
	overrideFileName           = "override_file"
	globalBucketName           = "global"
	beServicesPrefixRegex      = `be\/services`
	uatBucketName              = "uat"
	liveBucketPrefix           = "live_"
	uatBucketPrefix            = "uat_"
	localBucketName            = "local"
	filePrefix                 = "FILES/"
	configTypePropertiesString = "PROPERTIES"
	configTypeFileString       = "FILE"
)

type configType int

const (
	configTypeProperties = iota
	configTypeFile
)

type ConfigValue struct {
	Value  string `json:"value"`
	Bucket string `json:"bucket"`
}

type ConfigFile struct {
	Type       string                  `json:"type"`
	Contents   *ConfigValue            `json:"contents,omitempty"`
	Properties map[string]*ConfigValue `json:"properties,omitempty"`
}

func GetAllConfigBuckets(configBucket string, schema map[string]map[string]interface{}, service string, isLocal bool) ([]string, error) {
	var allConfigBuckets []string
	if isLocal {
		allConfigBuckets = []string{localBucketName}
	} else {
		allConfigBuckets = []string{}
	}

	allConfigBuckets = addConfigBucket(configBucket, service, allConfigBuckets)

	var currentBucket = configBucket

	for currentBucket != globalBucketName {
		if currentBucket == uatBucketName {
			currentBucket = liveBucketPrefix + configBucket[len(uatBucketPrefix):]
		} else {
			schemaEntry := schema[currentBucket]
			parents := schemaEntry["parents"].([]interface{})
			parent := parents[0].(string)

			currentBucket = parent
		}

		allConfigBuckets = addConfigBucket(currentBucket, service, allConfigBuckets)
	}

	return allConfigBuckets, nil
}

func addConfigBucket(bucket, service string, allConfigBuckets []string) []string {
	return append(allConfigBuckets, bucket+"#"+service, bucket)
}

func GetConfigs(buckets []string, configs map[string]string, tokensWithValues map[string]string) (map[string]ConfigValue, error) {
	configsByFileName := make(map[string]ConfigValue)

	for i := len(buckets) - 1; i >= 0; i = i - 1 {
		bucket := buckets[i]

		r, err := regexp.Compile(beServicesPrefixRegex + `\/[^\/]*\/` + bucket + `\/(.*)`)
		if err != nil {
			return nil, err
		}

		for k, v := range configs {
			fileName := r.FindStringSubmatch(k)
			if fileName != nil && fileName[1] != "" {
				configsByFileName[fileName[1]] = ConfigValue{
					Value:  base64.RawStdEncoding.EncodeToString([]byte(tokenizeConfigValue(v, tokensWithValues))),
					Bucket: bucket,
				}
			}
		}
	}

	return configsByFileName, nil
}

func tokenizeConfigValue(configValue string, tokensWithValues map[string]string) string {
	var newConfigValue = configValue
	for k, v := range tokensWithValues {
		tokenString := fmt.Sprintf("${%s}", k)
		newConfigValue = strings.ReplaceAll(newConfigValue, tokenString, v)
	}

	return newConfigValue
}

func GroupConfigs(configs map[string]ConfigValue) map[string]ConfigFile {
	result := make(map[string]ConfigFile)
	for k, v := range configs {
		// this is necessary because we want to use a pointer to v, and golang uses the same pointer for each loop iteration.
		value := v

		fileName, configType := getFileName(k)
		if configType == configTypeProperties {
			configFile := result[fileName]
			var properties map[string]*ConfigValue
			if configFile.Properties != nil {
				properties = configFile.Properties
			} else {
				properties = make(map[string]*ConfigValue)
				configFile.Type = configTypePropertiesString
			}

			configKey := getConfigKey(k)

			properties[configKey] = &value
			configFile.Properties = properties

			result[fileName] = configFile
		} else {
			result[fileName] = ConfigFile{
				Type:     configTypeFileString,
				Contents: &value,
			}
		}
	}

	return result
}

func getFileName(fileNameAndKey string) (string, configType) {
	if strings.HasPrefix(fileNameAndKey, filePrefix) {
		return fileNameAndKey[len(filePrefix):], configTypeFile
	}

	return strings.Split(fileNameAndKey, "#")[0], configTypeProperties
}

func getConfigKey(fileNameAndKey string) string {
	if strings.Contains(fileNameAndKey, "#") {
		return strings.Split(fileNameAndKey, "#")[1]
	}

	return fileNameAndKey
}
