package commands

import (
	"fmt"
)

// PlaceholderCommand represents a command scheduled for upcoming PRDs.
type PlaceholderCommand struct {
	name        string
	description string
	usage       string
	targetPRD   string
}

// NewPlaceholderCommand creates a placeholder command entry.
func NewPlaceholderCommand(name, description, usage, targetPRD string) Command {
	return &PlaceholderCommand{
		name:        name,
		description: description,
		usage:       usage,
		targetPRD:   targetPRD,
	}
}

func (p *PlaceholderCommand) Name() string {
	return p.name
}

func (p *PlaceholderCommand) Description() string {
	return p.description
}

func (p *PlaceholderCommand) Usage() string {
	return p.usage
}

func (p *PlaceholderCommand) Execute(ctx *Context, args []string) error {
	ctx.Printer.Info(fmt.Sprintf("Command '%s' is registered as a placeholder.", p.name))
	ctx.Printer.Println(fmt.Sprintf("Description: %s", p.description))
	ctx.Printer.Println(fmt.Sprintf("Scheduled for implementation in: %s", p.targetPRD))
	return nil
}

// RegisterPlaceholders adds all placeholder commands defined in PRD 001 to the registry.
func RegisterPlaceholders(r *Registry) {
	placeholders := []PlaceholderCommand{
		{
			name:        "inspect",
			description: "Inspect local or remote NodePhone runtime state",
			usage:       "nodephone inspect",
			targetPRD:   "Future PRD",
		},
	}

	for _, p := range placeholders {
		cmd := p // copy loop variable
		r.Register(&cmd)
	}
}
