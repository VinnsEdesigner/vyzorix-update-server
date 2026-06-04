package controllers

// WebSocket upgrade handling lives on Server.stream in server.go. The upgraded
// gorilla connection is handed to hub.Client, which owns readPump/writePump.
