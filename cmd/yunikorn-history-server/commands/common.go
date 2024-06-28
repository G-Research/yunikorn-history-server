package commands

import "fmt"

func validate() error {
	if ConfigFile == "" {
		return fmt.Errorf("config file is required")
	}
	return nil
}
