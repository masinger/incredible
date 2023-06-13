package logging

import (
	"github.com/masinger/incredible/internal/interactive"
	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger
var Interactive interactive.Interactive
