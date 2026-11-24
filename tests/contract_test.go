package contract_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	agentv1 "github.com/relayx-tech/contracts/gen/go/agent/v1"
	corev1 "github.com/relayx-tech/contracts/gen/go/core/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

// repoRoot 返回 contracts 仓库根目录（测试运行于 tests/ 子目录）。
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("无法定位测试文件")
	}
	return filepath.Dir(filepath.Dir(file))
}

func readFile(t *testing.T, rel string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", rel, err)
	}
	return b
}

// assertMethods 断言服务描述符暴露的方法集合与期望完全一致。
func assertMethods(t *testing.T, name string, desc grpc.ServiceDesc, want ...string) {
	t.Helper()
	got := make(map[string]bool, len(desc.Methods))
	for _, m := range desc.Methods {
		got[m.MethodName] = true
	}
	if len(got) != len(want) {
		t.Fatalf("%s 方法数 = %d, 期望 %d (%v)", name, len(got), len(want), want)
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("%s 缺少方法 %q", name, w)
		}
	}
}

func TestCoreServiceDescriptors(t *testing.T) {
	assertMethods(t, "ConfigService", corev1.ConfigService_ServiceDesc, "Validate", "Generate", "GenerateBatch", "Preview")
	assertMethods(t, "TunnelService", corev1.TunnelService_ServiceDesc, "GenerateTunnel")
	assertMethods(t, "ToolsService", corev1.ToolsService_ServiceDesc, "ScanReality", "GenerateX25519", "ValidateCertificate")
	assertMethods(t, "VersionService", corev1.VersionService_ServiceDesc, "Info", "Capabilities")
	assertMethods(t, "HealthService", corev1.HealthService_ServiceDesc, "Check")
	assertMethods(t, "NodeRuntimeService", corev1.NodeRuntimeService_ServiceDesc,
		"Apply", "Rollback", "KernelStart", "KernelStop", "KernelReload", "KernelStatus", "Perception")
}

func TestAgentServiceDescriptor(t *testing.T) {
	assertMethods(t, "AgentControlService", agentv1.AgentControlService_ServiceDesc,
		"Hello", "Apply", "Rollback", "Status", "Perception", "Interfaces", "Probe", "Version", "Health", "Update")
}

func TestRegistrationServiceDescriptor(t *testing.T) {
	assertMethods(t, "RegistrationService", agentv1.RegistrationService_ServiceDesc, "Register", "Renew", "GetCA")
}

func TestRegisterMessages(t *testing.T) {
	req := &agentv1.RegisterRequest{
		Meta:     &agentv1.AgentRequestMeta{OperationId: "op-1"},
		Token:    "token-1",
		NodeId:   "node-1",
		NodeType: agentv1.AgentNodeType_AGENT_NODE_TYPE_ENTRY,
		CsrPem:   []byte("csr"),
	}
	if req.GetToken() != "token-1" || req.GetNodeId() != "node-1" || len(req.GetCsrPem()) == 0 {
		t.Fatal("RegisterRequest 字段缺失")
	}
	resp := &agentv1.RegisterResponse{
		CertChainPem:                []byte("cert"),
		CaPem:                       []byte("ca"),
		NotAfter:                    timestamppb.Now(),
		PinnedCertFingerprintSha256: "fp",
	}
	if resp.GetPinnedCertFingerprintSha256() != "fp" || resp.GetNotAfter() == nil {
		t.Fatal("RegisterResponse 字段缺失")
	}
	if len((&agentv1.GetCAResponse{CaPem: []byte("ca")}).GetCaPem()) == 0 {
		t.Fatal("GetCAResponse 字段缺失")
	}
}

func TestNodeConfigSpecRoundTrip(t *testing.T) {
	spec := &corev1.NodeConfigSpec{
		NodeId:            "node-entry-1",
		NodeType:          corev1.NodeType_NODE_TYPE_ENTRY,
		SchemaVersion:     "1.0.0",
		TargetXrayVersion: "26.11.9",
		PortGroups: []*corev1.PortGroupConfig{
			{
				PortGroupId: "pg-1",
				Name:        "web",
				Inbounds: []*corev1.Inbound{
					{
						Tag:      "in-pg-1-443",
						Listen:   "0.0.0.0",
						Port:     443,
						Protocol: corev1.Protocol_PROTOCOL_VLESS,
						Sniffing: &corev1.Sniffing{Enabled: true, DestinationOverride: []string{"tls"}, DomainsExcluded: []string{"example.com"}},
					},
				},
				Clients: []*corev1.Client{{Email: "a@b.c", Id: "uuid", Flow: "xtls-rprx-vision", Level: 0}},
				TransportSecurity: &corev1.TransportSecurity{
					Transport: corev1.Transport_TRANSPORT_RAW,
					Security:  corev1.Security_SECURITY_TLS,
					Reality:   &corev1.RealitySettings{Dest: "www.example.com:443", ServerNames: []string{"www.example.com"}},
				},
				Outbounds:   []*corev1.Outbound{{Tag: "out-direct", Protocol: corev1.OutboundProtocol_OUTBOUND_PROTOCOL_FREEDOM}},
				Routing:     &corev1.Routing{DefaultOutboundTag: "out-direct", Rules: []*corev1.RoutingRule{{Domain: []string{"geosite:cn"}, OutboundTag: "out-direct"}}},
				Dns:         &corev1.Dns{Mode: corev1.DnsMode_DNS_MODE_REMOTE, Family: corev1.AddressFamily_ADDRESS_FAMILY_IPV4_PREFER, Servers: []string{"1.1.1.1"}},
				Policy:      &corev1.Policy{Levels: []*corev1.LevelPolicy{{Level: 0, ConnIdleSeconds: 300}}},
				FaultAction: corev1.FaultAction_FAULT_ACTION_REJECT,
			},
		},
		Materials: []*corev1.Material{{MaterialId: "m-1", Kind: corev1.MaterialKind_MATERIAL_KIND_TLS_CERT, FingerprintSha256: "abc"}},
	}

	raw, err := proto.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got corev1.NodeConfigSpec
	if err := proto.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !proto.Equal(spec, &got) {
		t.Fatalf("round-trip 不一致")
	}
}

func TestRequestMetaAndErrorDetail(t *testing.T) {
	meta := &corev1.RequestMeta{
		RequestId:       "req-1",
		OperationId:     "op-1",
		ContractVersion: "1.0.0",
		Timestamp:       timestamppb.Now(),
		Nonce:           "nonce-1234567890abcdef",
		Actor:           "panel",
	}
	if meta.GetRequestId() != "req-1" {
		t.Fatal("RequestMeta.RequestId 缺失")
	}

	detail := &corev1.ErrorDetail{
		Code:      "VALIDATION_INVALID_PORT",
		Field:     "port_groups[0].inbounds[0].port",
		RequestId: "req-1",
		Retryable: false,
		Message:   "端口不合法",
		Details:   []*corev1.ErrorItem{{Field: "port_groups[0].inbounds[0].port", Code: "OUT_OF_RANGE", Message: "端口须在 2000-30000"}},
	}
	if detail.GetDetails()[0].GetCode() != "OUT_OF_RANGE" {
		t.Fatal("ErrorDetail.Details 缺失")
	}
}

func TestApplyRollbackMessages(t *testing.T) {
	apply := &corev1.ApplyRequest{
		Meta:                &corev1.RequestMeta{OperationId: "op-1"},
		NodeId:              "node-entry-1",
		Version:             42,
		ConfigHash:          "hash",
		ConfigJson:          "{}",
		SchemaVersion:       "1.0.0",
		XrayVersion:         "26.11.9",
		RestoredFromVersion: 41,
	}
	if apply.GetVersion() != 42 {
		t.Fatal("ApplyRequest.Version 缺失")
	}
	rb := &corev1.RollbackRequest{Meta: &corev1.RequestMeta{}, NodeId: "n", TargetVersion: 40, NewVersion: 43, ConfigHash: "h", ConfigJson: "{}", Reason: "manual"}
	if rb.GetNewVersion() != 43 {
		t.Fatal("RollbackRequest.NewVersion 缺失")
	}
	ks := &corev1.KernelState{State: "RUNNING", Pid: 1234, Restarts: 1, XrayVersion: "26.11.9"}
	if ks.GetLastError() != "" && ks.State != "RUNNING" {
		t.Fatal("KernelState 字段异常")
	}
}

func TestAgentMessages(t *testing.T) {
	hello := &agentv1.HelloRequest{
		Meta:            &agentv1.AgentRequestMeta{RequestId: "r", ContractVersion: "1.0.0"},
		PanelVersion:    "1.0.0",
		AgentVersion:    "1.0.0",
		ContractVersion: "1.0.0",
	}
	if hello.GetPanelVersion() != "1.0.0" {
		t.Fatal("HelloRequest.PanelVersion 缺失")
	}
	apply := &agentv1.AgentApplyRequest{NodeId: "n", Version: 7, ConfigHash: "h", ConfigJson: "{}", SchemaVersion: "1.0.0", XrayVersion: "26.11.9"}
	if apply.GetVersion() != 7 {
		t.Fatal("AgentApplyRequest.Version 缺失")
	}
	ni := &agentv1.NetworkInterface{Name: "eth0", Mac: "aa:bb", Mtu: 1500, Up: true, Addresses: []string{"1.2.3.4/24"}}
	if ni.GetName() != "eth0" {
		t.Fatal("NetworkInterface.Name 缺失")
	}
}

func TestEnumsCoverDesignMatrix(t *testing.T) {
	// 入站 10 类
	inbound := []string{"PROTOCOL_VMESS", "PROTOCOL_VLESS", "PROTOCOL_TROJAN", "PROTOCOL_SHADOWSOCKS", "PROTOCOL_SOCKS", "PROTOCOL_HTTP", "PROTOCOL_WIREGUARD", "PROTOCOL_TUNNEL", "PROTOCOL_HYSTERIA", "PROTOCOL_MASQUE"}
	for _, name := range inbound {
		if _, ok := corev1.Protocol_value[name]; !ok {
			t.Errorf("Protocol 缺少 %s", name)
		}
	}
	transports := []string{"TRANSPORT_RAW", "TRANSPORT_XHTTP", "TRANSPORT_MKCP", "TRANSPORT_GRPC", "TRANSPORT_WEBSOCKET", "TRANSPORT_HTTPUPGRADE", "TRANSPORT_HYSTERIA", "TRANSPORT_MASQUE"}
	for _, name := range transports {
		if _, ok := corev1.Transport_value[name]; !ok {
			t.Errorf("Transport 缺少 %s", name)
		}
	}
	securities := []string{"SECURITY_NONE", "SECURITY_TLS", "SECURITY_REALITY"}
	for _, name := range securities {
		if _, ok := corev1.Security_value[name]; !ok {
			t.Errorf("Security 缺少 %s", name)
		}
	}
}

type errorCode struct {
	Code       string `yaml:"code"`
	HTTP       int    `yaml:"http"`
	GRPC       string `yaml:"grpc"`
	I18N       string `yaml:"i18n"`
	Owner      string `yaml:"owner"`
	Since      string `yaml:"since"`
	Deprecated bool   `yaml:"deprecated"`
	Message    string `yaml:"message"`
}

func TestErrorCodeRegistry(t *testing.T) {
	var codes []errorCode
	if err := yaml.Unmarshal(readFile(t, "errors/v1/error-codes.yaml"), &codes); err != nil {
		t.Fatalf("解析错误码注册表失败: %v", err)
	}
	if len(codes) == 0 {
		t.Fatal("错误码注册表为空")
	}
	byCode := make(map[string]errorCode, len(codes))
	for _, c := range codes {
		if _, dup := byCode[c.Code]; dup {
			t.Errorf("错误码重复: %s", c.Code)
		}
		if c.Code == "" || c.HTTP == 0 || c.GRPC == "" || c.Owner == "" || c.Since == "" || c.Message == "" {
			t.Errorf("错误码 %q 字段不完整", c.Code)
		}
		if c.I18N != "errors."+c.Code {
			t.Errorf("错误码 %q 的 i18n = %q, 期望 errors.%s", c.Code, c.I18N, c.Code)
		}
		byCode[c.Code] = c
	}
	// design §3.4 目录必须全部登记
	for _, want := range []string{
		"OK", "VALIDATION_MISSING_FIELD", "VALIDATION_INVALID_PORT", "VALIDATION_UNSUPPORTED_COMBINATION",
		"VALIDATION_MATERIAL_MISSING", "CONFIG_INVALID", "CONFIG_HASH_MISMATCH", "CONFIG_TEST_FAILED",
		"CONFIG_APPLY_FAILED", "CONFIG_ROLLBACK_FAILED", "CONFIG_TIMEOUT", "CONFIG_TOO_LARGE",
		"CONFIG_VERSION_STALE", "OPERATION_CONFLICT", "OPERATION_IN_PROGRESS", "VERSION_INCOMPATIBLE",
		"AUTH_UNAUTHENTICATED", "AUTH_SIGNATURE_INVALID", "AUTH_REPLAY_REJECTED", "AUTH_CERT_REVOKED",
		"NODE_NOT_FOUND", "NODE_TYPE_MISMATCH", "PORT_CONFLICT", "TUNNEL_MAPPING_INVALID",
		"TUNNEL_PORT_CONFLICT", "INTERFACE_NOT_FOUND", "KERNEL_NOT_RUNNING", "KERNEL_START_FAILED",
		"KERNEL_RELOAD_FAILED", "KERNEL_UNHEALTHY", "PERCEPTION_UNAVAILABLE", "PROBE_FAILED",
		"PROBE_TIMEOUT", "RATE_LIMITED", "INTERNAL",
	} {
		if _, ok := byCode[want]; !ok {
			t.Errorf("错误码注册表缺少 %s", want)
		}
	}
	// gRPC status 语义抽检
	if byCode["VERSION_INCOMPATIBLE"].GRPC != "FAILED_PRECONDITION" || byCode["VERSION_INCOMPATIBLE"].HTTP != 426 {
		t.Error("VERSION_INCOMPATIBLE 映射错误")
	}
	if byCode["OPERATION_CONFLICT"].GRPC != "ALREADY_EXISTS" || byCode["OPERATION_CONFLICT"].HTTP != 409 {
		t.Error("OPERATION_CONFLICT 映射错误")
	}
}

func TestCompatMatrix(t *testing.T) {
	var m map[string]string
	if err := yaml.Unmarshal(readFile(t, "compat/v1/compat.yaml"), &m); err != nil {
		t.Fatalf("解析兼容矩阵失败: %v", err)
	}
	for _, k := range []string{"contract", "core", "schema", "agent", "xray"} {
		if strings.TrimSpace(m[k]) == "" {
			t.Errorf("兼容矩阵缺少键 %s", k)
		}
	}
	if m["xray"] != "26.11.9" {
		t.Errorf("xray 固定版本 = %q, 期望 26.11.9", m["xray"])
	}
}
