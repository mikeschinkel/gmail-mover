package cliutil

// FlagDef defines a command flag declaratively
type FlagDef struct {
	Name     string
	Default  any
	Usage    string
	Required bool
	String   *string
	Bool     *bool
	Int64    *int64
}

func (fd *FlagDef) Type() (ft FlagType) {
	switch {
	case fd.String != nil:
		return StringFlag
	case fd.Bool != nil:
		return BoolFlag
	case fd.Int64 != nil:
		return Int64Flag
	}
	return UnknownFlagType
}

func (fd *FlagDef) SetValue(value any) {
	switch fd.Type() {
	case StringFlag:
		v := *value.(*string)
		if fd.String != nil {
			*fd.String = v
		}
	case BoolFlag:
		v := *value.(*bool)
		if fd.Bool != nil {
			*fd.Bool = v
		}
	case Int64Flag:
		v := *value.(*int64)
		if fd.Int64 != nil {
			*fd.Int64 = v
		}
	case UnknownFlagType:
		// Just here to have all flag types in the switch
	}
}
