package controllers

// Command controller handlers live on Server in server.go: command,
// requireHMAC, and authorizeDashboardOrHMAC. Offline command wakeups are
// delegated through services/fcm.Notifier when the WebSocket hub cannot deliver
// directly.
