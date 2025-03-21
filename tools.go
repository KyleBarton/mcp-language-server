package main

import (
	"fmt"

	"github.com/isaacphi/mcp-language-server/internal/tools"
	mcp_golang "github.com/metoro-io/mcp-golang"
)

type ReadDefinitionArgs struct {
	SymbolName      string `json:"symbolName" jsonschema:"required,description=The name of the symbol whose definition you want to find (e.g. 'mypackage.MyFunction', 'MyType.MyMethod')"`
	ShowLineNumbers bool   `json:"showLineNumbers" jsonschema:"required,default=true,description=Include line numbers in the returned source code"`
}

type FindReferencesArgs struct {
	SymbolName      string `json:"symbolName" jsonschema:"required,description=The name of the symbol to search for (e.g. 'mypackage.MyFunction', 'MyType')"`
	ShowLineNumbers bool   `json:"showLineNumbers" jsonschema:"required,default=true,description=Include line numbers when showing where the symbol is used"`
}

type ApplyTextEditArgs struct {
	FilePath string           `json:"filePath"`
	Edits    []tools.TextEdit `json:"edits"`
}

type GetDiagnosticsArgs struct {
	FilePath        string `json:"filePath" jsonschema:"required,description=The path to the file to get diagnostics for"`
	IncludeContext  bool   `json:"includeContext" jsonschema:"default=false,description=Include additional context for each diagnostic. Prefer false."`
	ShowLineNumbers bool   `json:"showLineNumbers" jsonschema:"required,default=true,description=If true, adds line numbers to the output"`
}

type GetCodeLensArgs struct {
	FilePath string `json:"filePath" jsonschema:"required,description=The path to the file to get code lens information for"`
}

type ExecuteCodeLensArgs struct {
	FilePath string `json:"filePath" jsonschema:"required,description=The path to the file containing the code lens to execute"`
	Index    int    `json:"index" jsonschema:"required,description=The index of the code lens to execute (from get_codelens output), 1 indexed"`
}

type Prompt struct {
	Symbol string `json:"symbol"`
}

func (s *server) registerPrompts() error {
	err := s.mcpServer.RegisterPrompt("read-definition", "call the read_definition tool TODO", func(argument Prompt) (*mcp_golang.PromptResponse, error) {
		text, err := tools.ReadDefinition(s.ctx, s.lspClient, argument.Symbol, true)
		if err != nil {
			return mcp_golang.NewPromptResponse(
					"Error response",
					mcp_golang.NewPromptMessage(mcp_golang.NewTextContent(fmt.Sprintf("There is an error reading the definition for %s", argument.Symbol)), mcp_golang.RoleUser)),
				nil
		}
		return mcp_golang.NewPromptResponse(
				"Read definition prompt response",
				mcp_golang.NewPromptMessage(mcp_golang.NewTextContent(text), mcp_golang.RoleUser)),
			nil
	})

	if err != nil {
		return fmt.Errorf("failed to register read-definition prompt: %v", err)
	}

	err = s.mcpServer.RegisterPrompt("find-references", "call the find_references tool TODO", func(argument Prompt) (*mcp_golang.PromptResponse, error) {
		text, err := tools.FindReferences(s.ctx, s.lspClient, argument.Symbol, true)
		if err != nil {
			return mcp_golang.NewPromptResponse(
					"Error response",
					mcp_golang.NewPromptMessage(mcp_golang.NewTextContent(fmt.Sprintf("There is an error finding references for %s", argument.Symbol)), mcp_golang.RoleUser)),
				nil
		}
		return mcp_golang.NewPromptResponse(
				"Find references prompt response",
				mcp_golang.NewPromptMessage(mcp_golang.NewTextContent(text), mcp_golang.RoleUser)),
			nil
	})

	if err != nil {
		return fmt.Errorf("failed to register find-references prompt: %v", err)
	}

	err = s.mcpServer.RegisterPrompt("get-codelens", "call the get_codelens tool TODO", func(argument Prompt) (*mcp_golang.PromptResponse, error) {
		text, err := tools.GetCodeLens(s.ctx, s.lspClient, argument.Symbol)
		if err != nil {
			return mcp_golang.NewPromptResponse(
					"Error response",
					mcp_golang.NewPromptMessage(mcp_golang.NewTextContent(fmt.Sprintf("There is an error getting codelens for file %s", argument.Symbol)), mcp_golang.RoleUser)),
				nil
		}
		return mcp_golang.NewPromptResponse(
				"Get codelens prompt response",
				mcp_golang.NewPromptMessage(mcp_golang.NewTextContent(text), mcp_golang.RoleUser)),
			nil
	})
	if err != nil {
		return fmt.Errorf("failed to register get-codelens prompt: %v", err)
	}

	err = s.mcpServer.RegisterPrompt("get-diagnostics", "call the get_diagnostics tool TODO", func(argument Prompt) (*mcp_golang.PromptResponse, error) {
		text, err := tools.GetDiagnosticsForFile(s.ctx, s.lspClient, argument.Symbol, true, true)
		if err != nil {
			return mcp_golang.NewPromptResponse(
					"Error response",
					mcp_golang.NewPromptMessage(mcp_golang.NewTextContent(fmt.Sprintf("There is an error getting diagnostics for file %s", argument.Symbol)), mcp_golang.RoleUser)),
				nil
		}
		return mcp_golang.NewPromptResponse(
				"Get diagnostics prompt response",
				mcp_golang.NewPromptMessage(mcp_golang.NewTextContent(text), mcp_golang.RoleUser)),
			nil
	})

	if err != nil {
		return fmt.Errorf("failed to register get-diagnostics prompt: %v", err)
	}

	return nil

}

func (s *server) registerTools() error {

	err := s.mcpServer.RegisterTool(
		"apply_text_edit",
		"Apply multiple text edits to a file.",
		func(args ApplyTextEditArgs) (*mcp_golang.ToolResponse, error) {
			response, err := tools.ApplyTextEdits(s.ctx, s.lspClient, args.FilePath, args.Edits)
			if err != nil {
				return nil, fmt.Errorf("Failed to apply edits: %v", err)
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(response)), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register tool: %v", err)
	}

	err = s.mcpServer.RegisterTool(
		"read_definition",
		"Read the source code definition of a symbol (function, type, constant, etc.) from the codebase. Returns the complete implementation code where the symbol is defined.",
		func(args ReadDefinitionArgs) (*mcp_golang.ToolResponse, error) {
			text, err := tools.ReadDefinition(s.ctx, s.lspClient, args.SymbolName, args.ShowLineNumbers)
			if err != nil {
				return nil, fmt.Errorf("Failed to get definition: %v", err)
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(text)), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register tool: %v", err)
	}

	err = s.mcpServer.RegisterTool(
		"find_references",
		"Find all usages and references of a symbol throughout the codebase. Returns a list of all files and locations where the symbol appears.",
		func(args FindReferencesArgs) (*mcp_golang.ToolResponse, error) {
			text, err := tools.FindReferences(s.ctx, s.lspClient, args.SymbolName, args.ShowLineNumbers)
			if err != nil {
				return nil, fmt.Errorf("Failed to find references: %v", err)
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(text)), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register tool: %v", err)
	}

	err = s.mcpServer.RegisterTool(
		"get_diagnostics",
		"Get diagnostic information for a specific file from the language server.",
		func(args GetDiagnosticsArgs) (*mcp_golang.ToolResponse, error) {
			text, err := tools.GetDiagnosticsForFile(s.ctx, s.lspClient, args.FilePath, args.IncludeContext, args.ShowLineNumbers)
			if err != nil {
				return nil, fmt.Errorf("Failed to get diagnostics: %v", err)
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(text)), nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to register tool: %v", err)
	}

	err = s.mcpServer.RegisterTool(
		"get_codelens",
		"Get code lens hints for a given file from the language server.",
		func(args GetCodeLensArgs) (*mcp_golang.ToolResponse, error) {
			text, err := tools.GetCodeLens(s.ctx, s.lspClient, args.FilePath)
			if err != nil {
				return nil, fmt.Errorf("Failed to get code lens: %v", err)
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(text)), nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to register tool: %v", err)
	}

	err = s.mcpServer.RegisterTool(
		"execute_codelens",
		"Execute a code lens command for a given file and lens index.",
		func(args ExecuteCodeLensArgs) (*mcp_golang.ToolResponse, error) {
			text, err := tools.ExecuteCodeLens(s.ctx, s.lspClient, args.FilePath, args.Index)
			if err != nil {
				return nil, fmt.Errorf("Failed to execute code lens: %v", err)
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(text)), nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to register tool: %v", err)
	}

	return nil
}
