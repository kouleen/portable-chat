package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultWhitelistPath = "configs/mcjoin_whitelist.txt"
	defaultProtocol      = 760
)

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

// IpToUint32 ip字符串转uint32
func IpToUint32(ipStr string) (uint32, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, fmt.Errorf("invalid ip")
	}
	ip = ip.To4()
	if ip == nil {
		return 0, fmt.Errorf("not ipv4")
	}
	return binary.BigEndian.Uint32(ip), nil
}

// Uint32ToIp uint32转回ipv4字符串
func Uint32ToIp(n uint32) string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, n)
	return net.IPv4(b[0], b[1], b[2], b[3]).String()
}

func main() {
	startIpStr := "45.0.0.1"
	endIpStr := "45.255.255.254"
	ipToUint32, err := IpToUint32(endIpStr)
	if err != nil {
		panic(err)
	}
	var toUint32 uint32 = 0
	for toUint32 < ipToUint32 {
		toUint32, err = IpToUint32(startIpStr)
		if err != nil {
			panic(err)
		}
		ip := Uint32ToIp(toUint32 + 1)
		startIpStr = ip
		aaaab := ip + ":25565"
		aaaa(aaaab)
	}

}

func aaaa(targetFlag string) {

	hostFlag := ""
	portFlag := 0
	usernameFlag := "LocalJoinTesg"
	timeoutFlag := 4 * time.Second
	protocolFlag := 0

	target, err := resolveTarget(targetFlag, hostFlag, portFlag, flag.Args())
	if err != nil {
		exitWithError(err)
		return
	}
	if err := validateUsername(usernameFlag); err != nil {
		exitWithError(err)
		return
	}

	protocol := protocolFlag
	versionName := "unknown"
	if protocol <= 0 {
		status, statusErr := queryJavaStatus(target, timeoutFlag)
		if statusErr != nil {
			exitWithError(fmt.Errorf("failed to auto-detect protocol: %w; you can retry with --protocol", statusErr))
			return
		}
		protocol = status.Version.Protocol
		versionName = status.Version.Name
		fmt.Printf("status: detected java server %s (protocol %d), players %d/%d\n",
			status.Version.Name,
			status.Version.Protocol,
			status.Players.Online,
			status.Players.Max,
		)
	}
	if protocol <= 0 {
		protocol = defaultProtocol
	}

	result, loginErr := loginSmokeTest(target, usernameFlag, protocol, timeoutFlag)
	if loginErr != nil {
		exitWithError(loginErr)
		return
	}

	fmt.Println("result: login smoke test completed")
	fmt.Printf("target: %s\n", target)
	fmt.Printf("username: %s\n", usernameFlag)
	fmt.Printf("protocol: %d\n", protocol)
	if versionName != "unknown" {
		fmt.Printf("version: %s\n", versionName)
	}
	fmt.Printf("stage: %s\n", result.Stage)
	fmt.Printf("detail: %s\n", result.Detail)
}

type LoginResult struct {
	Stage  string
	Detail string
}

func loginSmokeTest(target, username string, protocol int, timeout time.Duration) (*LoginResult, error) {
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		return nil, err
	}
	port, _ := strconv.Atoi(portStr)

	handshakeBody := &bytes.Buffer{}
	writeVarInt(handshakeBody, 0x00)
	writeVarInt(handshakeBody, protocol)
	writeString(handshakeBody, host)
	if err := binary.Write(handshakeBody, binary.BigEndian, uint16(port)); err != nil {
		return nil, err
	}
	writeVarInt(handshakeBody, 2)
	if err := writePacket(conn, handshakeBody.Bytes()); err != nil {
		return nil, err
	}

	loginBody := &bytes.Buffer{}
	writeVarInt(loginBody, 0x00)
	writeString(loginBody, username)
	if protocol >= 759 {
		writeBool(loginBody, false)
	}
	if err := writePacket(conn, loginBody.Bytes()); err != nil {
		return nil, err
	}

	compressionEnabled := false
	for i := 0; i < 6; i++ {
		packetID, payload, readErr := readMinecraftPacket(conn, compressionEnabled)
		if readErr != nil {
			return nil, readErr
		}
		switch packetID {
		case 0x00:
			reason, _ := readString(bytes.NewReader(payload))
			return &LoginResult{
				Stage:  "login_disconnect",
				Detail: chatToPlain(reason),
			}, nil
		case 0x01:
			if protocol >= 764 {
				msgID, _ := readVarInt(bytes.NewReader(payload))
				channel, _ := readString(bytes.NewReader(payload))
				return &LoginResult{
					Stage:  "login_plugin_request",
					Detail: fmt.Sprintf("server requested login plugin negotiation (message id %d, channel %s)", msgID, channel),
				}, nil
			}
			return &LoginResult{
				Stage:  "encryption_request",
				Detail: "server requires authenticated online-mode encryption; this tool stops before Mojang/Microsoft auth",
			}, nil
		case 0x02:
			return &LoginResult{
				Stage:  "login_success",
				Detail: "login success packet received; the server accepted the single-player login probe",
			}, nil
		case 0x03:
			if protocol >= 764 {
				return &LoginResult{
					Stage:  "login_ack_required",
					Detail: "login success acknowledged stage reached on a modern server; probe stops before configuration/play traffic",
				}, nil
			}
			compressionEnabled = true
		case 0x04:
			if protocol >= 764 {
				return &LoginResult{
					Stage:  "cookie_request",
					Detail: "server requested a login cookie; probe stops here by design",
				}, nil
			}
			return &LoginResult{
				Stage:  "login_plugin_request",
				Detail: "server requested login plugin negotiation; probe stops here by design",
			}, nil
		default:
			return &LoginResult{
				Stage:  "unexpected_packet",
				Detail: fmt.Sprintf("received packet id 0x%02x during login", packetID),
			}, nil
		}
	}

	return &LoginResult{
		Stage:  "timeout_waiting_for_login_result",
		Detail: "no definitive login response was received before the probe stopped",
	}, nil
}

func queryJavaStatus(target string, timeout time.Duration) (*JavaStatus, error) {
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		return nil, err
	}
	port, _ := strconv.Atoi(portStr)

	handshakeBody := &bytes.Buffer{}
	writeVarInt(handshakeBody, 0x00)
	writeVarInt(handshakeBody, defaultProtocol)
	writeString(handshakeBody, host)
	if err := binary.Write(handshakeBody, binary.BigEndian, uint16(port)); err != nil {
		return nil, err
	}
	writeVarInt(handshakeBody, 1)
	if err := writePacket(conn, handshakeBody.Bytes()); err != nil {
		return nil, err
	}
	if err := writePacket(conn, []byte{0x00}); err != nil {
		return nil, err
	}

	packetLen, readErr := readVarInt(conn)
	if readErr != nil {
		return nil, readErr
	}
	packetData := make([]byte, packetLen)
	if _, readErr := io.ReadFull(conn, packetData); readErr != nil {
		return nil, readErr
	}

	reader := bytes.NewReader(packetData)
	packetID, readErr := readVarInt(reader)
	if readErr != nil {
		return nil, readErr
	}
	if packetID != 0x00 {
		return nil, fmt.Errorf("unexpected status packet id %d", packetID)
	}

	jsonLen, readErr := readVarInt(reader)
	if readErr != nil {
		return nil, readErr
	}
	jsonData := make([]byte, jsonLen)
	if _, readErr := io.ReadFull(reader, jsonData); readErr != nil {
		return nil, readErr
	}

	var status JavaStatus
	if unmarshalErr := json.Unmarshal(jsonData, &status); unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return &status, nil
}

func ensureTargetAllowed(host, whitelistPath string) error {
	if host == "" {
		return errors.New("empty host")
	}
	trimmed := strings.Trim(host, "[]")
	if strings.EqualFold(trimmed, "localhost") {
		return nil
	}

	if ip := net.ParseIP(trimmed); ip != nil && ip.IsLoopback() {
		return nil
	}

	whitelist, err := loadWhitelist(whitelistPath)
	if err != nil {
		return err
	}
	if _, ok := whitelist[strings.ToLower(trimmed)]; ok {
		return nil
	}

	return fmt.Errorf("target host %q is blocked; only localhost/loopback is allowed by default, or add it to %s for your own test servers", trimmed, whitelistPath)
}

func loadWhitelist(path string) (map[string]struct{}, error) {
	allowed := make(map[string]struct{})
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return allowed, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		allowed[strings.ToLower(line)] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return allowed, nil
}

func validateUsername(username string) error {
	if len(username) == 0 || len(username) > 16 {
		return errors.New("username must be 1-16 characters")
	}
	for _, r := range username {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return errors.New("username may only contain letters, numbers, and underscore")
	}
	return nil
}

func readMinecraftPacket(r io.Reader, compressed bool) (int, []byte, error) {
	packetLen, err := readVarInt(r)
	if err != nil {
		return 0, nil, err
	}
	packetData := make([]byte, packetLen)
	if _, err := io.ReadFull(r, packetData); err != nil {
		return 0, nil, err
	}

	payload := packetData
	if compressed {
		reader := bytes.NewReader(packetData)
		dataLen, dataLenErr := readVarInt(reader)
		if dataLenErr != nil {
			return 0, nil, dataLenErr
		}
		if dataLen != 0 {
			zr, zlibErr := zlib.NewReader(reader)
			if zlibErr != nil {
				return 0, nil, zlibErr
			}
			defer zr.Close()
			payload, zlibErr = io.ReadAll(zr)
			if zlibErr != nil {
				return 0, nil, zlibErr
			}
		} else {
			payload, dataLenErr = io.ReadAll(reader)
			if dataLenErr != nil {
				return 0, nil, dataLenErr
			}
		}
	}

	payloadReader := bytes.NewReader(payload)
	packetID, err := readVarInt(payloadReader)
	if err != nil {
		return 0, nil, err
	}
	rest, err := io.ReadAll(payloadReader)
	if err != nil {
		return 0, nil, err
	}
	return packetID, rest, nil
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

func writeBool(w io.Writer, value bool) {
	if value {
		_, _ = w.Write([]byte{1})
		return
	}
	_, _ = w.Write([]byte{0})
}

func readString(r io.Reader) (string, error) {
	length, err := readVarInt(r)
	if err != nil {
		return "", err
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return "", err
	}
	return string(data), nil
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

func chatToPlain(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "server closed the login with an empty reason"
	}

	var payload any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return raw
	}
	return flattenChat(payload)
}

func flattenChat(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case map[string]any:
		var parts []string
		if text, ok := v["text"].(string); ok {
			parts = append(parts, text)
		}
		if translate, ok := v["translate"].(string); ok && translate != "" {
			parts = append(parts, translate)
		}
		if extra, ok := v["extra"].([]any); ok {
			for _, item := range extra {
				parts = append(parts, flattenChat(item))
			}
		}
		return strings.TrimSpace(strings.Join(parts, ""))
	case []any:
		var parts []string
		for _, item := range v {
			parts = append(parts, flattenChat(item))
		}
		return strings.TrimSpace(strings.Join(parts, ""))
	default:
		data, _ := json.Marshal(v)
		return string(data)
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

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	fmt.Fprintf(os.Stderr, "tip: only localhost/loopback is allowed by default; add your own test host to %s if needed\n", filepath.Clean(defaultWhitelistPath))
}
