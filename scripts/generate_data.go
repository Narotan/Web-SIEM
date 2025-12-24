package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Event struct {
	ID        string `json:"_id"`
	AgentID   string `json:"agent_id"`
	Timestamp string `json:"timestamp"`
	BatchTime string `json:"batch_time"`
	EventType string `json:"event_type"`
	Severity  string `json:"severity"`
	User      string `json:"user"`
	Process   string `json:"process"`
	Source    string `json:"source"`
	SourceIP  string `json:"source_ip"`
	Hostname  string `json:"hostname"`
	Message   string `json:"message"`
	Command   string `json:"command"`
	RawLog    string `json:"raw_log"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	eventTypes := []string{"user_login", "auth_failure", "system_event", "file_access", "network_connection", "process_start", "privilege_escalation", "config_change"}
	severities := []string{"low", "medium", "high", "critical"}
	users := []string{"admin", "root", "developer", "operator", "security", "guest", "service_account", "backup_user", "nginx", "postgres"}
	processes := []string{"sshd", "sudo", "systemd", "nginx", "postgres", "docker", "cron", "auditd", "firewalld", "journald"}
	agents := []string{"agent-ubuntu-01", "agent-centos-02", "agent-debian-03", "agent-arch-04", "agent-fedora-05"}
	sources := []string{"auth", "syslog", "audit", "kernel", "application", "security"}
	hostnames := []string{"webserver-01", "dbserver-02", "appserver-03", "monitoring-04", "gateway-05"}

	// Load existing data or create new map
	data := make(map[string]Event)

	existingFile := "NoSQLdb/data/security_events.json"
	if fileData, err := os.ReadFile(existingFile); err == nil {
		json.Unmarshal(fileData, &data)
		fmt.Printf("Loaded %d existing events\n", len(data))
	}

	now := time.Now()

	// Generate 150 events over the last 24 hours with good distribution
	for i := 0; i < 150; i++ {
		// Spread events across last 24 hours
		hoursAgo := rand.Intn(24)
		minutesAgo := rand.Intn(60)
		secondsAgo := rand.Intn(60)

		eventTime := now.Add(-time.Duration(hoursAgo)*time.Hour - time.Duration(minutesAgo)*time.Minute - time.Duration(secondsAgo)*time.Second)

		eventType := eventTypes[rand.Intn(len(eventTypes))]
		severity := severities[rand.Intn(len(severities))]
		user := users[rand.Intn(len(users))]
		process := processes[rand.Intn(len(processes))]
		agent := agents[rand.Intn(len(agents))]
		source := sources[rand.Intn(len(sources))]
		hostname := hostnames[rand.Intn(len(hostnames))]

		// Weight severities - more low/medium, fewer high/critical
		sevRand := rand.Intn(100)
		if sevRand < 40 {
			severity = "low"
		} else if sevRand < 75 {
			severity = "medium"
		} else if sevRand < 92 {
			severity = "high"
		} else {
			severity = "critical"
		}

		ip := fmt.Sprintf("192.168.%d.%d", rand.Intn(256), rand.Intn(256))
		port := 1024 + rand.Intn(64000)

		var message, command string
		switch eventType {
		case "user_login":
			message = fmt.Sprintf("Accepted password for %s from %s port %d ssh2", user, ip, port)
			command = message
		case "auth_failure":
			message = fmt.Sprintf("Failed password for %s from %s port %d ssh2", user, ip, port)
			command = message
			severity = "high" // Auth failures are always high
		case "system_event":
			message = fmt.Sprintf("System event: service %s status changed to running", process)
			command = fmt.Sprintf("systemctl status %s", process)
		case "file_access":
			files := []string{"/etc/passwd", "/etc/shadow", "/var/log/auth.log", "/etc/ssh/sshd_config", "/root/.ssh/authorized_keys"}
			message = fmt.Sprintf("File access: %s read %s", user, files[rand.Intn(len(files))])
			command = fmt.Sprintf("cat %s", files[rand.Intn(len(files))])
		case "network_connection":
			message = fmt.Sprintf("New connection from %s to port %d via %s", ip, port, process)
			command = fmt.Sprintf("connect %s:%d", ip, port)
		case "process_start":
			message = fmt.Sprintf("Process %s (PID %d) started by %s", process, 1000+rand.Intn(50000), user)
			command = fmt.Sprintf("/usr/bin/%s --daemon", process)
		case "privilege_escalation":
			message = fmt.Sprintf("%s escalated privileges via sudo to execute /bin/bash", user)
			command = fmt.Sprintf("%s : TTY=pts/0 ; PWD=/home/%s ; USER=root ; COMMAND=/bin/bash", user, user)
			severity = "high"
		case "config_change":
			configs := []string{"/etc/sysconfig/network", "/etc/nginx/nginx.conf", "/etc/ssh/sshd_config", "/etc/firewalld/zones/public.xml"}
			config := configs[rand.Intn(len(configs))]
			message = fmt.Sprintf("Configuration file %s modified by %s", config, user)
			command = fmt.Sprintf("vim %s", config)
		}

		id := fmt.Sprintf("%d-%d", eventTime.UnixNano(), rand.Intn(1000000))

		rawLog := fmt.Sprintf("%s %s %s[%d]: %s",
			eventTime.Format("Jan 02 15:04:05"),
			hostname,
			process,
			1000+rand.Intn(50000),
			message)

		event := Event{
			ID:        id,
			AgentID:   agent,
			Timestamp: eventTime.Format(time.RFC3339),
			BatchTime: now.Format(time.RFC3339),
			EventType: eventType,
			Severity:  severity,
			User:      user,
			Process:   process,
			Source:    source,
			SourceIP:  ip,
			Hostname:  hostname,
			Message:   message,
			Command:   command,
			RawLog:    rawLog,
		}

		data[id] = event
	}

	// Write back to file
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	err = os.WriteFile(existingFile, output, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Printf("Successfully generated data. Total events: %d\n", len(data))

	// Print summary
	typeCounts := make(map[string]int)
	sevCounts := make(map[string]int)
	for _, e := range data {
		typeCounts[e.EventType]++
		sevCounts[e.Severity]++
	}

	fmt.Println("\nEvent types:")
	for t, c := range typeCounts {
		fmt.Printf("  %s: %d\n", t, c)
	}

	fmt.Println("\nSeverity distribution:")
	for s, c := range sevCounts {
		fmt.Printf("  %s: %d\n", s, c)
	}
}
