package functions

import (
	"fmt"
	"github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas"
	"log"
	"os"
	"strings"
	"text/tabwriter"
)

// "text/tabwriter"

func printFunctionList() {
	var groups = map[string]map[string]map[string]interface{}{}
	for index, schema := range schemas.RawSchemas {
		var group map[string]map[string]interface{}

		if groupName, ok := schema["group"].(string); ok {
			if thisGroup, ok := groups[groupName]; ok {
				group = thisGroup
			} else {
				group = map[string]map[string]interface{}{}
				groups[groupName] = group
			}
		} else {
			log.Fatalf("The schema for function %s does not have a group property.", index)
		}

		group[index] = schema
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	for groupName, group := range groups {
		fmt.Printf("\n%s\n---------------------------------\n\n", groupName)

		for index, schema := range group {
			fmt.Fprintf(writer, "%s\t%s\n", index, schema["description"])
		}

		writer.Flush()

		fmt.Println()
	}
}

func printFunctionHelp(name string) {
	name = strings.ToLower(name)

	if !strings.HasPrefix(name, "$") {
		name = "$" + name
	}

	schema, ok := schemas.RawSchemas[name]

	if !ok {
		fmt.Printf("Function `%s` not found.\n\n", name)
		return
	}

	fmt.Printf("Function `%s` - %s\n\n", schema["title"], schema["description"])

	if returnInfo, ok := schema["return"].(map[string]interface{}); ok {
		fmt.Printf("Returns (%s) %s\n\n", returnInfo["type"], returnInfo["description"])

		if returnInfo["type"].(string) == "object" {
			fmt.Println("\nReturned object properties\n--------------------------\n")

			writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			fmt.Fprintln(writer, "Name\tType\tDescription")
			fmt.Fprintln(writer, "----\t----\t-----------")

			props := returnInfo["properties"].(map[string]interface{})

			for index, prop := range props {
				prop := prop.(map[string]interface{})

				fmt.Fprintf(writer, "%s\t%s\t%s\n", index, prop["type"], prop["description"])
			}

			writer.Flush()
		}
	}

	fmt.Println("\nArguments\n---------\n")

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	requireds := schema["required"].([]interface{})
	properties := schema["properties"].(map[string]interface{})

	if len(properties) == 0 {
		fmt.Println("This function has no parameters.")
	} else {
		fmt.Fprintln(writer, "R\tName\tType\tDescription")
		fmt.Fprintln(writer, "-\t----\t----\t-----------")
	}

	for name, data := range properties {
		data := data.(map[string]interface{})
		required := ""

		for _, property := range requireds {
			if property == name {
				required = "*"
				break
			}
		}

		description, ok := data["description"]

		if !ok {
			description = "No description available"
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n", required, name, data["type"], description)
	}

	writer.Flush()

	fmt.Println()
}

func PrintHelp(functionName string) {
	if functionName == "" {
		printFunctionList()
		return
	}

	printFunctionHelp(functionName)
}
