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
type DatabaseObject struct {
	id             int
	name           string
	created        int
	deleted        int
	createRevision int
	prevRevision   int
	lease          int
	value          string
	oldValue       string
}

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
		fmt.Println("up called")

		apiResourcesDir := "/Users/dn/Documents/logs/tickets/beta-test/bundle-20200217T015050/cluster-data/api-resources"

		var files []string
		// Find yaml files in directory -- Maybe clean this up a bit in the future
		err := filepath.Walk(apiResourcesDir, func(path string, info os.FileInfo, err error) error {
			if filepath.Ext(path) == ".yaml" {
				files = append(files, path)
				// basenames = append(files, info.Name())
				return nil
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}

		// Set base resourceID to something out of range
		var resourceID int = 100000000

		// For each yaml file in apiresources dir
		for _, file := range files {
			// Get the file basename from its full path
			basename := file[strings.LastIndex(file, "/")+1:]

			// Get api resource group from file name
			apiResourceGroup := strings.TrimRight(strings.SplitN(basename, ".", 2)[1], ".yaml")

			// Write out api resource groups that break path in database
			switch apiResourceGroup {
			case "apps", "certificates.k8s.io", "coordination.k8s.io", "extensions", "networking.k8s.io", "rbac.authorization.k8s.io", "scheduling.k8s.io", "storage.k8s.io", "snapshot.storage.k8s.io":
				apiResourceGroup = ""
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
				objectJSON, err := json.Marshal(kubernetesItems.Items[i])
				if err != nil {
					log.Fatal(err)
				}

				// Save the raw JSON []byte so that we can unmarshal it later on and also preserve it raw
				// fullObjectJSON := objectJSON

				var kubernetesObject KubernetesObject

				// Unmarshal json to kubernetes object struct to write as sql statements
				err = json.Unmarshal(objectJSON, &kubernetesObject)
				if err != nil {
					log.Fatal(err)
				}

				// Filter name and namespace
				// name := kubernetesObject.Metadata.Name
				// fmt.Println("Name: ", name)
				// if kubernetesObject.Metadata.Namespace != "" {
				// 	namespace := kubernetesObject.Metadata.Namespace
				// 	fmt.Println("Namespace: ", namespace)
				// }

				// Print the full object json
				// fmt.Printf("%s\n\n", fullObjectJSON)

				// var sqlInsertStatement string
				// var apiResourceName string
				// var apiResourceNamespaced bool
				// var objectState []byte
				// var objectName string
				// var objectNamespace string

				// objectState = fullObjectJSON
				// objectName = kubernetesObject.Metadata.Name

				// INSERT INTO kine(id, name, created, deleted, create_revision, prev_revision, lease, value, old_value) VALUES($RESOURCE_ID, '/registry/$API_RESOURCE_GROUP/$API_RESOURCE_NAME/$NAMESPACE/$ITEMNAME', 1, 0, $((RESOURCE_ID + 1)), $((RESOURCE_ID + 2)), 0, '$ITEMSTATE', '$ITEMSTATE');

				sqlInsertStatement := fmt.Sprintf(
					"INSERT INTO kine(id, name, created, deleted, create_revision, prev_revision, lease, value, old_value)"+
						"VALUES(%d, '/registry/\n",
					resourceID)

				resourceID++

				// if

				fmt.Printf("%s", sqlInsertStatement)
			}
		}
	},
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
