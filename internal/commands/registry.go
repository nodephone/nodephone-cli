package commands

import (
	"fmt"
	"sort"
	"sync"
)

// Registry manages registered CLI commands.
type Registry struct {
	mu       sync.RWMutex
	commands map[string]Command
	order    []string
}

// NewRegistry creates a new command registry instance.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register registers a command into the registry.
func (r *Registry) Register(cmd Command) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := cmd.Name()
	if _, exists := r.commands[name]; !exists {
		r.order = append(r.order, name)
	}
	r.commands[name] = cmd
}

// Get retrieves a command by name.
func (r *Registry) Get(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmd, exists := r.commands[name]
	return cmd, exists
}

// List returns registered commands in registration/insertion order or sorted order.
func (r *Registry) List() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]Command, 0, len(r.order))
	for _, name := range r.order {
		if cmd, exists := r.commands[name]; exists {
			list = append(list, cmd)
		}
	}
	return list
}

// SortedList returns registered commands sorted by name.
func (r *Registry) SortedList() []Command {
	list := r.List()
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name() < list[j].Name()
	})
	return list
}

// Execute looks up and executes a command by name.
func (r *Registry) Execute(ctx *Context, name string, args []string) error {
	cmd, exists := r.Get(name)
	if !exists {
		return fmt.Errorf("unknown command %q. Run 'nodephone help' for available commands", name)
	}
	return cmd.Execute(ctx, args)
}
