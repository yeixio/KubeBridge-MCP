package server

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// InvocationResult holds the result of invoking a tool handler
type InvocationResult struct {
	CallToolResult *mcp.CallToolResult
	Output         interface{}
}

// invokeToolHandler invokes a tool handler using reflection
// The handler is expected to be of type mcp.ToolHandlerFor[In, Out]
func invokeToolHandler(ctx context.Context, handler interface{}, args map[string]interface{}, toolName string) (*InvocationResult, error) {
	if handler == nil {
		return nil, fmt.Errorf("handler is nil")
	}

	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()

	// Verify it's a function
	if handlerType.Kind() != reflect.Func {
		return nil, fmt.Errorf("handler is not a function")
	}

	// Verify it has the right signature: func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error)
	if handlerType.NumIn() != 3 || handlerType.NumOut() != 3 {
		return nil, fmt.Errorf("handler has wrong signature")
	}

	// Get the input type (3rd parameter)
	inputType := handlerType.In(2)

	// Create a new instance of the input type
	inputValue := reflect.New(inputType).Elem()

	// Fill the input struct from the args map
	if err := fillStructFromMap(inputValue, args); err != nil {
		return nil, fmt.Errorf("failed to fill input struct: %w", err)
	}

	// Create a minimal CallToolRequest for the handler
	// We'll create a pointer to a zero-value CallToolRequest
	callToolRequestType := reflect.TypeOf((*mcp.CallToolRequest)(nil)).Elem()
	callReqPtr := reflect.New(callToolRequestType)
	callReqValue := callReqPtr.Elem()

	// Try to set common fields if they exist
	if nameField := callReqValue.FieldByName("Name"); nameField.IsValid() && nameField.CanSet() {
		if toolName != "" {
			nameField.SetString(toolName)
		} else {
			nameField.SetString("dynamic_call")
		}
	}

	// Call the handler: handler(ctx, callReq, inputValue)
	// Pass the pointer, not the value
	results := handlerValue.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		callReqPtr,
		inputValue,
	})

	// Check for error (3rd return value)
	if !results[2].IsNil() {
		err := results[2].Interface().(error)
		return nil, err
	}

	// Extract results
	callToolResult := results[0].Interface().(*mcp.CallToolResult)
	outputValue := results[1].Interface()

	return &InvocationResult{
		CallToolResult: callToolResult,
		Output:         outputValue,
	}, nil
}

// fillStructFromMap fills a struct from a map using reflection
func fillStructFromMap(structValue reflect.Value, m map[string]interface{}) error {
	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}

	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		// Get JSON tag name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Remove options from JSON tag (e.g., "name,omitempty" -> "name")
		jsonName := strings.Split(jsonTag, ",")[0]

		// Get value from map
		if value, ok := m[jsonName]; ok {
			if field.CanSet() {
				valueReflect := reflect.ValueOf(value)
				if valueReflect.Type().AssignableTo(field.Type()) {
					field.Set(valueReflect)
				} else if field.Type().Kind() == reflect.String && valueReflect.Kind() == reflect.String {
					field.Set(valueReflect)
				} else {
					// Try to convert via JSON
					valueJSON, err := json.Marshal(value)
					if err == nil {
						if err := json.Unmarshal(valueJSON, field.Addr().Interface()); err == nil {
							continue
						}
					}
					// Last resort: try direct assignment if types are compatible
					if valueReflect.Type().ConvertibleTo(field.Type()) {
						field.Set(valueReflect.Convert(field.Type()))
					}
				}
			}
		}
	}

	return nil
}
