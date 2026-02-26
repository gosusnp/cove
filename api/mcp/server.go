package mcp

import (
	"net/http"

	"github.com/gosusnp/cove/api/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewHTTPHandler(exercises *service.ExerciseService) http.Handler {
	server := mcp.NewServer(&mcp.Implementation{Name: "cove", Version: "1.0.0"}, nil)
	registerExerciseTools(server, exercises)
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, nil)
}
