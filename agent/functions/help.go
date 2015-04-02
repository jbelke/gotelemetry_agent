package functions

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas"
	"io"
	"log"
	"os"
	"strings"
)

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

	for groupName, group := range groups {
		writer := tablewriter.NewWriter(os.Stdout)

		fmt.Printf("\n%s\n---------------------------------\n\n", groupName)

		for index, schema := range group {
			writer.Append([]string{index, schema["description"].(string)})
		}

		writer.Render()

		fmt.Println()
	}
}

func printArgList(schema map[string]interface{}, output io.Writer) {
	if output == nil {
		output = os.Stdout
	}

	writer := tablewriter.NewWriter(output)

	requireds, ok := schema["required"].([]interface{})

	if !ok {
		requireds = []interface{}{}
	}

	properties, ok := schema["properties"].(map[string]interface{})

	if !ok {
		properties = map[string]interface{}{"@self": schema}
	}

	if len(properties) == 0 {
		fmt.Fprintln(output, "This function has no parameters.")
	} else {
		writer.SetHeader([]string{"Required", "Name", "Type", "Description"})
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

		description, ok := data["description"].(string)

		if !ok {
			description = "No description available"
		}

		typeName, ok := data["type"]

		if !ok {
			typeName = "--"
		}

		var possibilities []interface{}

		if anyOf, ok := data["anyOf"].([]interface{}); ok {
			possibilities = anyOf
			description += ". Can be any of:"
		}

		if oneOf, ok := data["oneOf"].([]interface{}); ok {
			possibilities = oneOf
			description += ". Must be one of:"
		}

		if len(possibilities) > 0 {
			description += "\n"

			for index, group := range possibilities {
				var b bytes.Buffer

				printArgList(group.(map[string]interface{}), &b)

				description += string(b.Bytes())

				if index < len(possibilities)-1 {
					description += "\nor\n\n"
				}
			}
		}

		writer.Append([]string{required, name, fmt.Sprintf("%v", typeName), description})
	}

	writer.Render()
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

			writer := tablewriter.NewWriter(os.Stdout)

			writer.SetHeader([]string{"Name", "Type", "Description"})

			props := returnInfo["properties"].(map[string]interface{})

			for index, prop := range props {
				propData := prop.(map[string]interface{})

				writer.Append([]string{index, fmt.Sprintf("%v", propData["type"]), fmt.Sprintf("%v", propData["description"])})
			}

			writer.Render()
		}
	}

	fmt.Println("\nArguments\n---------\n")

	printArgList(schema, nil)
	fmt.Println()
}

func PrintHelp(functionName string) {
	if functionName == "" {
		printFunctionList()
		return
	}

	printFunctionHelp(functionName)
}
