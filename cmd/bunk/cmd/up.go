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

		f, err := os.OpenFile("/Users/dn/Documents/logs/tickets/beta-test/bundle-20200217T015050/cluster-data/api-resources/nodes.yaml", os.O_RDONLY, 0444)

		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

		content, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal(err)
		}

		var kubernetesItems KubernetesItems
		err = yaml.Unmarshal(content, &kubernetesItems)
		if err != nil {
			log.Fatal(err)
		}

		for i := 0; i < len(kubernetesItems.Items); i++ {
			objectJSON, err := json.Marshal(kubernetesItems.Items[i])
			if err != nil {
				log.Fatal(err)
			}

			fullObjectJSON := objectJSON

			var kubernetesObject KubernetesObject

			err = json.Unmarshal(objectJSON, &kubernetesObject)
			if err != nil {
				log.Fatal(err)
			}

			name := kubernetesObject.Metadata.Name
			fmt.Println("Name: ", name)
			if kubernetesObject.Metadata.Namespace != "" {
				namespace := kubernetesObject.Metadata.Namespace
				fmt.Println("Namespace: ", namespace)
			}

			fmt.Printf("%s\n\n", fullObjectJSON)
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
