package main

import (
	"github.com/snowdrop/spring-boot-cloud-devex/cmd"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/common/logger"
)

func main() {
	// Enable Debug if env var is defined
	logger.EnableLogLevelDebug()

	// Call commands
	cmd.Execute()
}
