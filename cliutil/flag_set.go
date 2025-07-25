package cliutil

import (
	"errors"
	"flag"
	"fmt"
	"slices"
	"strings"
)

// FlagSet combines a FlagSet with automatic config binding
type FlagSet struct {
	Name     string
	FlagSet  *flag.FlagSet
	FlagDefs []FlagDef
	Values   map[string]any
}

// Parse extracts flags and returns remaining args
func (fs *FlagSet) Parse(args []string) (remainingArgs []string, err error) {
	var fsFlagNames, fsArgs, nonFSArgs []string

	if fs == nil {
		err = fmt.Errorf("FlagSet is nil")
		goto end
	}

	// Parse only the flags, collect non-flag arguments
	fsFlagNames = fs.FlagNames()

	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			// This is a non-flag argument (command name or command arg)
			nonFSArgs = append(nonFSArgs, arg)
			continue
		}
		if !slices.Contains(fsFlagNames, strings.TrimPrefix(arg, "-")) {
			// This is some other flag, leave it for command parsing
			nonFSArgs = append(nonFSArgs, arg)
			continue
		}
		// This is not a flagSet flag, add it to a temp slice for parsing
		fsArgs = append(fsArgs, arg)
	}

	if len(fsArgs) == 0 {
		goto end
	}

	err = fs.Build()
	if err != nil {
		goto end
	}

	// Parse the global flags we found
	err = fs.FlagSet.Parse(fsArgs)
	if err != nil {
		goto end
	}

	err = fs.Assign()

end:
	return nonFSArgs, err
}

func (fs *FlagSet) Build() (err error) {
	var errs []error

	if fs.Name == "" {
		err = fmt.Errorf("name cannot be empty for FlagSet with flags %v", fs.FlagNames())
	}

	fs.FlagSet = flag.NewFlagSet(fs.Name, flag.ContinueOnError)
	fs.Values = make(map[string]any)

	// Add all defined flags to the flag set
	for _, flagDef := range fs.FlagDefs {
		switch flagDef.Type() {
		case StringFlag:
			defaultVal := ""
			if flagDef.Default != nil {
				defaultVal = flagDef.Default.(string)
			}
			fs.Values[flagDef.Name] = fs.FlagSet.String(flagDef.Name, defaultVal, flagDef.Usage)
		case BoolFlag:
			defaultVal := false
			if flagDef.Default != nil {
				defaultVal = flagDef.Default.(bool)
			}
			fs.Values[flagDef.Name] = fs.FlagSet.Bool(flagDef.Name, defaultVal, flagDef.Usage)
		case Int64Flag:
			defaultVal := int64(0)
			if flagDef.Default != nil {
				defaultVal = flagDef.Default.(int64)
			}
			fs.Values[flagDef.Name] = fs.FlagSet.Int64(flagDef.Name, defaultVal, flagDef.Usage)
		default:
			errs = append(errs, fmt.Errorf("unknown flag type for %s", flagDef.Name))
		}
	}
	if len(errs) > 0 {
		err = errors.Join(errs...)
	}
	return err
}

func (fs *FlagSet) FlagNames() (names []string) {
	names = make([]string, len(fs.FlagDefs))
	for i, fd := range fs.FlagDefs {
		names[i] = fd.Name
	}
	return names
}

// ParseAndBind parses flags and automatically binds values to config
func (fs *FlagSet) ParseAndBind(args []string, _ Config) (_ []string, err error) {
	var errs []error

	err = fs.FlagSet.Parse(args)
	if err != nil {
		goto end
	}
	for _, flagDef := range fs.FlagDefs {
		switch {
		case flagDef.Bool != nil:
			*flagDef.Bool = true // CLAUDE: <== how to get the value?
		case flagDef.String != nil:
			*flagDef.String = "" // CLAUDE: <== how to get the value?
		case flagDef.Int64 != nil:
			*flagDef.Int64 = 9 // CLAUDE: <== how to get the value?
		default:
			errs = append(errs, fmt.Errorf("unknown flag type: %s", flagDef.Type))
		}
	}
	if len(errs) > 0 {
		err = errors.Join(errs...)
		goto end
	}
	args = fs.FlagSet.Args()
end:
	return args, err
}

func (fs *FlagSet) Assign() (err error) {
	var errs []error
	for _, flagDef := range fs.FlagDefs {
		switch flagDef.Type() {
		case StringFlag:
			value := fs.Values[flagDef.Name].(*string)
			*flagDef.String = *value
		case BoolFlag:
			value := fs.Values[flagDef.Name].(*bool)
			*flagDef.Bool = *value
		case Int64Flag:
			value := fs.Values[flagDef.Name].(*int64)
			*flagDef.Int64 = *value
		default:
			errs = append(errs, fmt.Errorf("unknown flag type for %s", flagDef.Name))
		}
	}
	if len(errs) > 0 {
		err = errors.Join(errs...)
	}
	return err
}
