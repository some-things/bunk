/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// KubernetesItems : Structure of Kubernetes yaml files -- make this more clear/redefine
type KubernetesItems struct {
	Items []interface{} `yaml:"items"`
}

// KubernetesObject : Generic minimal structure for Kubernetes objects
type KubernetesObject struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace,omitempty"`
	} `yaml:"metadata"`
}

// DatabaseObject : Structure of each item in the k3d SQLite3 backend database
// type DatabaseObject struct {
// 	id             int
// 	name           string
// 	created        int
// 	deleted        int
// 	createRevision int
// 	prevRevision   int
// 	lease          int
// 	value          string
// 	oldValue       string
// }

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("up called")

		writeKubernetesResources()
	},
}

func writeKubernetesResources() {
	apiResourcesDir := "/Users/dn/Documents/logs/tickets/beta-test/bundle-20200217T015050/cluster-data/api-resources"

	var files []string
	// Find yaml files in directory -- Maybe clean this up a bit in the future
	err := filepath.Walk(apiResourcesDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".yaml" {
			files = append(files, path)
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Set base resourceID to something out of range
	resourceID := 100000000

	// For each yaml file in apiresources dir
	for _, file := range files {
		// Get the file basename from its full path
		basename := file[strings.LastIndex(file, "/")+1:]

		// Get api resource group from file name
		apiResourceGroup := strings.TrimSuffix(strings.SplitN(basename, ".", 2)[1], ".yaml")

		// Write out api resource groups that break path in database
		switch apiResourceGroup {
		case "apps",
			"certificates.k8s.io",
			"coordination.k8s.io",
			"extensions",
			"networking.k8s.io",
			"rbac.authorization.k8s.io",
			"scheduling.k8s.io",
			"storage.k8s.io",
			"snapshot.storage.k8s.io":
			apiResourceGroup = ""
		}

		// Get api resource name from file name
		apiResourceName := strings.SplitN(basename, ".", 2)[0]

		// Modify api resource names for known inconsistencies
		switch apiResourceName {
		case "nodes":
			apiResourceName = "minions"
		case "endpoints":
			apiResourceName = "services/endpoints"
		case "services":
			apiResourceName = "services/specs"
		case "leases":
			apiResourceName = "leases/kube-node-lease"
		case "ingresses":
			apiResourceName = "ingress"
		case "podsecuritypolicies":
			apiResourceName = "podsecuritypolicy"
		}

		// Open file as read-only
		f, err := os.OpenFile(file, os.O_RDONLY, 0444)
		if err != nil {
			log.Fatal(err)
		}

		// Close file when finished
		defer f.Close()

		// Read the file
		content, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal(err)
		}

		// Unmarshal yaml to json
		var kubernetesItems KubernetesItems
		err = yaml.Unmarshal(content, &kubernetesItems)
		if err != nil {
			log.Fatal(err)
		}

		// For each item in the Kubernetes resource file
		for i := 0; i < len(kubernetesItems.Items); i++ {
			// Marshal the json items to individual []byte
			// objectJSON, err := json.Marshal(kubernetesItems.Items[i])
			// if err != nil {
			// 	log.Fatal(err)
			// }

			// enc := json.NewEncoder(os.Stdout)
			// enc.SetEscapeHTML(false)
			// err = enc.Encode(objectJSON)
			// if err != nil {
			// 	log.Fatal(err)
			// }

			buffer := &bytes.Buffer{}
			enc := json.NewEncoder(buffer)
			enc.SetEscapeHTML(false)
			err := enc.Encode(kubernetesItems.Items[i])
			if err != nil {
				log.Fatal(err)
			}

			objectJSON := buffer.Bytes()

			// Preserve the raw JSON []byte and trim newline for printing >:(
			objectState := []byte(strings.TrimSuffix(string(objectJSON), "\n"))

			var kubernetesObject KubernetesObject

			// Unmarshal json to kubernetes object struct to write as sql statements
			err = json.Unmarshal(objectJSON, &kubernetesObject)
			if err != nil {
				log.Fatal(err)
			}

			// Print the full object json
			// fmt.Printf("%s\n\n", objectState)

			var sqlInsertStatement string = ""
			// var apiResourceName string
			// var apiResourceNamespaced bool
			// var objectState []byte
			// var objectName string
			// var objectNamespace string

			objectName := kubernetesObject.Metadata.Name
			objectNamespace := kubernetesObject.Metadata.Namespace

			sqlInsertStatement = fmt.Sprintf(
				"INSERT INTO kine(id, name, created, deleted, create_revision, prev_revision, lease, value, old_value) "+
					"VALUES(%d, '/registry/",
				resourceID)

			// Add api resource group to path if it exists
			if apiResourceGroup != "" {
				sqlInsertStatement += fmt.Sprintf("%s/", apiResourceGroup)
			}

			// Add api resource name
			sqlInsertStatement += fmt.Sprintf("%s/", apiResourceName)

			// Add namespace if it exists
			if objectNamespace != "" {
				sqlInsertStatement += fmt.Sprintf("%s/", objectNamespace)
			}

			// Escape objectState s/\'/\'\'/g as string and convert back to byte
			objectState = []byte(strings.ReplaceAll(string(objectState), "'", "''"))

			// Add required additional fields
			sqlInsertStatement += fmt.Sprintf("%s', 1, 0, %d, %d, 0, '%s', '%s');", objectName, resourceID+1, resourceID+2, objectState, objectState)

			// Increment resourceID
			// TODO: I think 3 is sufficient -- need to test this
			resourceID += 4

			fmt.Printf("%s\n", sqlInsertStatement)
		}
	}
}

func init() {
	rootCmd.AddCommand(upCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
