package controllers

// OTA controller handlers live on Server in server.go: version, changelog,
// apk, bin, serveDownload, and serveStatic. They remain methods so route
// wiring can share Config-driven asset roots and CORS/logging wrappers.
