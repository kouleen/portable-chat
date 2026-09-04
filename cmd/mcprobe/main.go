package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var bedrockMagic = []byte{0x00, 0xff, 0xff, 0x00, 0xfe, 0xfe, 0xfe, 0xfe, 0xfd, 0xfd, 0xfd, 0xfd, 0x12, 0x34, 0x56, 0x78}

type JavaStatus struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Online int `json:"online"`
		Max    int `json:"max"`
	} `json:"players"`
	Description any `json:"description"`
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic: %v", r)
		}
	}()
	for {
		timeoutFlag := 3 * time.Second

		//target := "mcfkcs.com:12345"
		target := "minecraft.kouleen.cn:25565"

		mode := strings.ToLower(strings.TrimSpace("auto"))
		switch mode {
		case "auto", "java", "bedrock":
		default:
			fmt.Fprintln(os.Stderr, "error: mode must be one of auto, java, bedrock")
			os.Exit(1)
		}

		var javaResult string
		var bedrockResult string

		if mode == "auto" || mode == "java" {
			if result, err := probeJava(target, timeoutFlag); err == nil {
				fmt.Println("result: minecraft java server")
				fmt.Println(result)
				continue
			} else {
				javaResult = err.Error()
			}
		}

		if mode == "auto" || mode == "bedrock" {
			if result, err := probeBedrock(target, timeoutFlag); err == nil {
				fmt.Println("result: minecraft bedrock server")
				fmt.Println(result)
				continue
			} else {
				bedrockResult = err.Error()
			}
		}

		fmt.Println("result: not identified as a minecraft server")
		if javaResult != "" {
			fmt.Println("java probe:", javaResult)
		}
		if bedrockResult != "" {
			fmt.Println("bedrock probe:", bedrockResult)
		}
		//os.Exit(2)
	}
}

func resolveTarget(target, host string, port int, args []string) (string, error) {
	if target != "" {
		return normalizeTarget(target)
	}
	if host != "" && port > 0 {
		return net.JoinHostPort(host, strconv.Itoa(port)), nil
	}
	if len(args) == 1 {
		return normalizeTarget(args[0])
	}
	if len(args) == 2 {
		p, err := strconv.Atoi(args[1])
		if err != nil || p <= 0 || p > 65535 {
			return "", errors.New("invalid port")
		}
		return net.JoinHostPort(args[0], args[1]), nil
	}
	return "", errors.New("missing target, please provide ip:port")
}

func normalizeTarget(target string) (string, error) {
	if !strings.Contains(target, ":") {
		return "", errors.New("target must be in the form ip:port")
	}
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		if strings.Count(target, ":") == 1 {
			parts := strings.SplitN(target, ":", 2)
			host, port = parts[0], parts[1]
		} else {
			return "", err
		}
	}
	if host == "" {
		return "", errors.New("host is empty")
	}
	p, err := strconv.Atoi(port)
	if err != nil || p <= 0 || p > 65535 {
		return "", errors.New("invalid port")
	}
	return net.JoinHostPort(host, strconv.Itoa(p)), nil
}

func probeJava(target string, timeout time.Duration) (string, error) {
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		return "", err
	}
	port, _ := strconv.Atoi(portStr)

	handshakeBody := &bytes.Buffer{}
	writeVarInt(handshakeBody, 0x00)
	writeVarInt(handshakeBody, 760)
	writeString(handshakeBody, host)
	if err := binary.Write(handshakeBody, binary.BigEndian, uint16(port)); err != nil {
		return "", err
	}
	writeVarInt(handshakeBody, 1)

	if err := writePacket(conn, handshakeBody.Bytes()); err != nil {
		return "", err
	}
	if err := writePacket(conn, []byte{0x00}); err != nil {
		return "", err
	}

	packetLen, readErr := readVarInt(conn)
	if readErr != nil {
		return "", readErr
	}
	packetData := make([]byte, packetLen)
	if _, readErr := io.ReadFull(conn, packetData); readErr != nil {
		return "", readErr
	}

	reader := bytes.NewReader(packetData)
	packetID, readErr := readVarInt(reader)
	if readErr != nil {
		return "", readErr
	}
	if packetID != 0x00 {
		return "", fmt.Errorf("unexpected packet id %d", packetID)
	}

	jsonLen, readErr := readVarInt(reader)
	if readErr != nil {
		return "", readErr
	}
	jsonData := make([]byte, jsonLen)
	if _, readErr := io.ReadFull(reader, jsonData); readErr != nil {
		return "", readErr
	}

	var status JavaStatus
	if unmarshalErr := json.Unmarshal(jsonData, &status); unmarshalErr != nil {
		return "", fmt.Errorf("status payload is not valid minecraft json: %w", unmarshalErr)
	}

	return fmt.Sprintf(
		"target: %s\nedition: java\nversion: %s (protocol %d)\nplayers: %d/%d\ndescription: %s",
		target,
		status.Version.Name,
		status.Version.Protocol,
		status.Players.Online,
		status.Players.Max,
		stringifyDescription(status.Description),
	), nil
}

func probeBedrock(target string, timeout time.Duration) (string, error) {
	conn, err := net.DialTimeout("udp", target, timeout)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	packet := &bytes.Buffer{}
	packet.WriteByte(0x01)
	if err := binary.Write(packet, binary.BigEndian, time.Now().UnixMilli()); err != nil {
		return "", err
	}
	packet.Write(bedrockMagic)
	if err := binary.Write(packet, binary.BigEndian, uint64(0x1122334455667788)); err != nil {
		return "", err
	}

	if _, err := conn.Write(packet.Bytes()); err != nil {
		return "", err
	}

	response := make([]byte, 2048)
	n, readErr := conn.Read(response)
	if readErr != nil {
		return "", readErr
	}
	response = response[:n]
	if len(response) < 35 {
		return "", errors.New("response too short")
	}
	if response[0] != 0x1c {
		return "", fmt.Errorf("unexpected packet id 0x%02x", response[0])
	}
	if !bytes.Equal(response[17:33], bedrockMagic) {
		return "", errors.New("raknet magic mismatch")
	}

	dataLen := int(binary.BigEndian.Uint16(response[33:35]))
	if len(response) < 35+dataLen {
		return "", errors.New("invalid bedrock payload length")
	}
	payload := string(response[35 : 35+dataLen])
	parts := strings.Split(payload, ";")
	if len(parts) < 6 || strings.ToUpper(parts[0]) != "MCPE" {
		return "", errors.New("payload does not look like a bedrock motd")
	}

	version := field(parts, 3)
	online := field(parts, 4)
	maxPlayers := field(parts, 5)
	motd := field(parts, 1)
	gameMode := field(parts, 8)

	return fmt.Sprintf(
		"target: %s\nedition: bedrock\nversion: %s\nplayers: %s/%s\ngamemode: %s\nmotd: %s",
		target,
		version,
		online,
		maxPlayers,
		gameMode,
		motd,
	), nil
}

func writePacket(w io.Writer, body []byte) error {
	packet := &bytes.Buffer{}
	writeVarInt(packet, len(body))
	packet.Write(body)
	_, err := w.Write(packet.Bytes())
	return err
}

func writeString(w io.Writer, s string) {
	writeVarInt(w, len(s))
	_, _ = io.WriteString(w, s)
}

func writeVarInt(w io.Writer, value int) {
	u := uint32(value)
	for {
		if (u & ^uint32(0x7f)) == 0 {
			_, _ = w.Write([]byte{byte(u)})
			return
		}
		_, _ = w.Write([]byte{byte(u&0x7f) | 0x80})
		u >>= 7
	}
}

func readVarInt(r io.Reader) (int, error) {
	var numRead int
	var result int
	for {
		var b [1]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, err
		}
		value := int(b[0] & 0x7f)
		result |= value << (7 * numRead)

		numRead++
		if numRead > 5 {
			return 0, errors.New("varint too big")
		}
		if b[0]&0x80 == 0 {
			break
		}
	}
	return result, nil
}

func stringifyDescription(v any) string {
	switch value := v.(type) {
	case string:
		return value
	case map[string]any:
		if text, ok := value["text"].(string); ok && text != "" {
			return text
		}
		if extra, ok := value["extra"].([]any); ok {
			var parts []string
			for _, item := range extra {
				parts = append(parts, stringifyDescription(item))
			}
			return strings.TrimSpace(strings.Join(parts, ""))
		}
	case []any:
		var parts []string
		for _, item := range value {
			parts = append(parts, stringifyDescription(item))
		}
		return strings.TrimSpace(strings.Join(parts, ""))
	}
	raw, _ := json.Marshal(v)
	return string(raw)
}

func field(parts []string, idx int) string {
	if idx >= 0 && idx < len(parts) {
		return parts[idx]
	}
	return ""
}
