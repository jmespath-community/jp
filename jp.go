package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/jmespath-community/go-jmespath"
	"github.com/nwidger/jsoncolor"
	"github.com/urfave/cli"
)

const version = "1.1.0"

func main() {
	app := cli.NewApp()
	app.Name = "jp"
	app.Version = fmt.Sprintf("%s, community edition", version)
	app.Usage = "jp [<options>] <expression>"
	app.Author = ""
	app.Email = ""
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "compact, c",
			Usage: "Produce compact JSON output that omits nonessential whitespace.",
		},
		cli.StringFlag{
			Name:  "filename, f",
			Usage: "Read input JSON from a file instead of stdin.",
		},
		cli.StringFlag{
			Name:  "expr-file, e",
			Usage: "Read JMESPath expression from the specified file.",
		},
		cli.StringFlag{
			Name:  "color",
			Value: "auto",
			Usage: "Change the color setting (none, auto, always). auto is based on whether output is a tty.",
		},
		cli.BoolFlag{
			Name:   "unquoted, u",
			Usage:  "If the final result is a string, it will be printed without quotes.",
			EnvVar: "JP_UNQUOTED",
		},
		cli.BoolFlag{
			Name:  "stream, s",
			Usage: "Parse JSON elements until the input stream is exhausted (rather than just the first JSON).",
		},
		cli.BoolFlag{
			Name:  "ast",
			Usage: "Only print the AST of the parsed expression.  Do not rely on this output, only useful for debugging purposes.",
		},
	}
	app.Action = runMainAndExit

	app.Run(os.Args)
}

func runMainAndExit(c *cli.Context) {
	os.Exit(runMain(c))
}

func errMsg(msg string, a ...interface{}) int {
	fmt.Fprintf(os.Stderr, msg, a...)
	fmt.Fprintln(os.Stderr)
	return 1
}

func runMain(c *cli.Context) int {
	var expression string
	if c.String("expr-file") != "" {
		byteExpr, err := ioutil.ReadFile(c.String("expr-file"))
		expression = string(byteExpr)
		if err != nil {
			return errMsg("Error opening expression file: %s", err)
		}
	} else {
		if len(c.Args()) == 0 {
			return errMsg("Must provide at least one argument.")
		}
		expression = c.Args()[0]
	}
	switch c.String("color") {
	case "always":
		enableColor(true)
	case "auto":
		// this requests the default behaviour in the jsoncolor library
		// color output is enabled or disabled dynamically based on the
		// stdout's file descriptor referring to a terminal or not.
		// Additionally, if the NO_COLOR environment variable is set
		// (regardless of its value) color output will be disabled.
	case "never":
		enableColor(false)
	default:
		return errMsg("Invalid color specification. Must use always/auto/never")
	}
	if c.Bool("ast") {
		parser := jmespath.NewParser()
		parsed, err := parser.Parse(expression)
		if err != nil {
			if syntaxError, ok := err.(jmespath.SyntaxError); ok {
				return errMsg("%s\n%s\n",
					syntaxError,
					syntaxError.HighlightLocation())
			}
			return errMsg("%s", err)
		}
		fmt.Println("")
		fmt.Printf("%s\n", parsed)
		return 0
	}
	var jsonParser *json.Decoder
	if c.String("filename") != "" {
		f, err := os.Open(c.String("filename"))
		if err != nil {
			return errMsg("Error opening input file: %s", err)
		}
		jsonParser = json.NewDecoder(f)

	} else {
		jsonParser = json.NewDecoder(os.Stdin)
	}
	for {
		var input interface{}
		if err := jsonParser.Decode(&input); err == io.EOF {
			break
		} else if err != nil {
			errMsg("Error parsing input json: %s\n", err)
			return 2
		}
		result, err := jmespath.Search(expression, input)
		if err != nil {
			if syntaxError, ok := err.(jmespath.SyntaxError); ok {
				return errMsg("%s\n%s\n",
					syntaxError,
					syntaxError.HighlightLocation())
			}
			return errMsg("Error evaluating JMESPath expression: %s", err)
		}
		converted, isString := result.(string)
		if c.Bool("unquoted") && isString {
			os.Stdout.WriteString(converted)
		} else {
			var toJSON []byte
			if c.Bool("compact") {
				toJSON, err = json.Marshal(result)
			} else {
				toJSON, err = json.MarshalIndent(result, "", "  ")
			}
			if err != nil {
				errMsg("Error marshalling result to JSON: %s\n", err)
				return 3
			}
			os.Stdout.Write(toJSON)
		}
		os.Stdout.WriteString("\n")
		if !c.Bool("stream") {
			break
		}
	}
	return 0
}

func enableColor(enabled bool) {

	if enabled {
		jsoncolor.DefaultArrayColor.EnableColor()
		jsoncolor.DefaultColonColor.EnableColor()
		jsoncolor.DefaultCommaColor.EnableColor()
		jsoncolor.DefaultFalseColor.EnableColor()
		jsoncolor.DefaultFieldColor.EnableColor()
		jsoncolor.DefaultFieldQuoteColor.EnableColor()
		jsoncolor.DefaultNullColor.EnableColor()
		jsoncolor.DefaultNumberColor.EnableColor()
		jsoncolor.DefaultObjectColor.EnableColor()
		jsoncolor.DefaultSpaceColor.EnableColor()
		jsoncolor.DefaultStringColor.EnableColor()
		jsoncolor.DefaultStringQuoteColor.EnableColor()
		jsoncolor.DefaultTrueColor.EnableColor()

	} else {
		jsoncolor.DefaultArrayColor.DisableColor()
		jsoncolor.DefaultColonColor.DisableColor()
		jsoncolor.DefaultCommaColor.DisableColor()
		jsoncolor.DefaultFalseColor.DisableColor()
		jsoncolor.DefaultFieldColor.DisableColor()
		jsoncolor.DefaultFieldQuoteColor.DisableColor()
		jsoncolor.DefaultNullColor.DisableColor()
		jsoncolor.DefaultNumberColor.DisableColor()
		jsoncolor.DefaultObjectColor.DisableColor()
		jsoncolor.DefaultSpaceColor.DisableColor()
		jsoncolor.DefaultStringColor.DisableColor()
		jsoncolor.DefaultStringQuoteColor.DisableColor()
		jsoncolor.DefaultTrueColor.DisableColor()
	}
}
