package aserto_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/header"
	assrt "github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type connectionRecorder struct {
	address     string
	tlsConf     *tls.Config
	callerCreds credentials.PerRPCCredentials
	tenantID    string
	sessionID   string
	dialOptions []grpc.DialOption
}

func (d *connectionRecorder) Connect(
	address string,
	tlsConf *tls.Config,
	callerCreds credentials.PerRPCCredentials,
	tenantID, sessionID string,
	options []grpc.DialOption,
) (*grpc.ClientConn, error) {
	d.address = address
	d.tlsConf = tlsConf
	d.callerCreds = callerCreds
	d.tenantID = tenantID
	d.sessionID = sessionID
	d.dialOptions = options

	return &grpc.ClientConn{}, nil
}

func TestWithAddr(t *testing.T) {
	assert := assrt.New(t)

	recorder := &connectionRecorder{}
	options, err := aserto.NewConnectionOptions(aserto.WithAddr("address"))
	assert.NoError(err)

	aserto.InternalNewConnection(recorder.Connect, options) //nolint: errcheck

	assert.Equal("address", recorder.address)
}

func TestWithURL(t *testing.T) {
	assert := assrt.New(t)
	recorder := &connectionRecorder{}

	const URL = "https://server.com:123"
	svcURL, err := url.Parse(URL)
	assert.NoError(err)

	options, err := aserto.NewConnectionOptions(aserto.WithURL(svcURL))
	assert.NoError(err)
	aserto.InternalNewConnection(recorder.Connect, options) //nolint: errcheck

	assert.Equal(URL, recorder.address)
}

func TestAddrAndURL(t *testing.T) {
	assert := assrt.New(t)
	svcURL, err := url.Parse("https://server.com:123")
	assert.NoError(err)

	_, err = aserto.NewConnectionOptions(aserto.WithAddr("address"), aserto.WithURL(svcURL))
	assert.Error(err)
}

func TestWithInsecure(t *testing.T) {
	assert := assrt.New(t)
	recorder := &connectionRecorder{}
	options, err := aserto.NewConnectionOptions(aserto.WithInsecure(true))
	assert.NoError(err)
	aserto.InternalNewConnection(recorder.Connect, options) //nolint: errcheck

	assert.True(recorder.tlsConf.InsecureSkipVerify)
}

func TestWithTokenAuth(t *testing.T) {
	assert := assrt.New(t)
	recorder := &connectionRecorder{}
	options, err := aserto.NewConnectionOptions(aserto.WithTokenAuth("<token>"))
	assert.NoError(err)
	aserto.InternalNewConnection(recorder.Connect, options) //nolint: errcheck

	md, err := recorder.callerCreds.GetRequestMetadata(context.TODO())
	assert.NoError(err)

	token, ok := md["authorization"]
	assert.True(ok)
	assert.Equal("bearer <token>", token)
}

func TestWithBearerTokenAuth(t *testing.T) {
	assert := assrt.New(t)

	recorder := &connectionRecorder{}
	options, err := aserto.NewConnectionOptions(aserto.WithTokenAuth("bearer <token>"))
	assert.NoError(err)
	aserto.InternalNewConnection(recorder.Connect, options) //nolint: errcheck

	md, err := recorder.callerCreds.GetRequestMetadata(context.TODO())
	assert.NoError(err)

	token, ok := md["authorization"]
	assert.True(ok)
	assert.Equal("bearer <token>", token)
}

func TestWithAPIKey(t *testing.T) {
	assert := assrt.New(t)
	recorder := &connectionRecorder{}
	options, err := aserto.NewConnectionOptions(aserto.WithAPIKeyAuth("<apikey>"))
	assert.NoError(err)
	aserto.InternalNewConnection(recorder.Connect, options) //nolint: errcheck

	md, err := recorder.callerCreds.GetRequestMetadata(context.TODO())
	assert.NoError(err)

	token, ok := md["authorization"]
	assert.True(ok)
	assert.Equal("basic <apikey>", token)
}

func TestTokenAndAPIKey(t *testing.T) {
	_, err := aserto.NewConnectionOptions(aserto.WithAPIKeyAuth("<apikey>"), aserto.WithTokenAuth("<token>"))
	assrt.Error(t, err)
}

func TestWithTenantID(t *testing.T) {
	assert := assrt.New(t)
	recorder := &connectionRecorder{}
	options, err := aserto.NewConnectionOptions(aserto.WithTenantID("<tenantid>"))
	assert.NoError(err)

	conn, err := aserto.InternalNewConnection(recorder.Connect, options) //nolint: errcheck
	assert.NoError(err)

	assert.Equal("<tenantid>", recorder.tenantID)

	ctx := context.TODO()
	err = aserto.InternalUnary("<tenantid>", "")(
		ctx,
		"method",
		"request",
		"reply",
		conn,
		func(c context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(c)
			assert.True(ok)

			tenantID := md.Get("aserto-tenant-id")
			assert.Equal(1, len(tenantID), "request should contain tenant ID metadata field")
			assert.Equal("<tenantid>", tenantID[0], "tenant ID metadata should have the expected value")

			assert.Equal("method", method, "'method' parameter should be a passthrough")
			assert.Equal("request", req, "'request' parameter should be a passthrough")
			assert.Equal("reply", reply, "'reply' parameter should be a passthrough")

			return nil
		})
	assert.NoError(err)

	_, err = aserto.InternalStream("<tenantid>", "")(
		ctx,
		nil,
		conn,
		"method",
		func(
			c context.Context,
			desc *grpc.StreamDesc,
			cc *grpc.ClientConn,
			method string,
			opts ...grpc.CallOption,
		) (grpc.ClientStream, error) {
			md, ok := metadata.FromOutgoingContext(c)
			assert.True(ok)

			tenantID := md.Get("aserto-tenant-id")
			assert.Equal(1, len(tenantID), "request should contain tenant ID metadata field")
			assert.Equal("<tenantid>", tenantID[0], "tenant ID metadata should have the expected value")

			assert.Equal("method", method, "'method' parameter should be a passthrough")

			return nil, nil
		},
	)
	assert.NoError(err)
}

func TestWithSessionID(t *testing.T) {
	assert := assrt.New(t)
	recorder := &connectionRecorder{}
	options, err := aserto.NewConnectionOptions(aserto.WithSessionID("<sessionid>"))
	assert.NoError(err)

	conn, err := aserto.InternalNewConnection(recorder.Connect, options) //nolint: errcheck
	assert.NoError(err)

	assert.Equal("<sessionid>", recorder.sessionID)

	ctx := context.TODO()
	err = aserto.InternalUnary("", "<sessionid>")(
		ctx,
		"method",
		"request",
		"reply",
		conn,
		func(c context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(c)
			assert.True(ok)

			sessionID := md.Get("aserto-session-id")
			assert.Equal(1, len(sessionID), "request should contain session ID metadata field")
			assert.Equal("<sessionid>", sessionID[0], "session ID metadata should have the expected value")

			assert.Equal("method", method, "'method' parameter should be a passthrough")
			assert.Equal("request", req, "'request' parameter should be a passthrough")
			assert.Equal("reply", reply, "'reply' parameter should be a passthrough")

			return nil
		})
	assert.NoError(err)

	_, err = aserto.InternalStream("", "<sessionid>")(
		ctx,
		nil,
		conn,
		"method",
		func(
			c context.Context,
			desc *grpc.StreamDesc,
			cc *grpc.ClientConn,
			method string,
			opts ...grpc.CallOption,
		) (grpc.ClientStream, error) {
			md, ok := metadata.FromOutgoingContext(c)
			assert.True(ok)

			sessionID := md.Get(string(header.HeaderAsertoSessionID))
			assert.Equal(1, len(sessionID), "request should contain session ID metadata field")
			assert.Equal("<sessionid>", sessionID[0], "session ID metadata should have the expected value")

			assert.Equal("method", method, "'method' parameter should be a passthrough")

			return nil, nil
		},
	)
	assert.NoError(err)
}

const CertSubjectName = "Testing Inc."

func TestWithCACertPath(t *testing.T) {
	assert := assrt.New(t)
	tempdir := t.TempDir()
	caPath := fmt.Sprintf("%s/ca.pem", tempdir)

	file, err := os.OpenFile(caPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	assert.NoError(err, "Failed to create CA file")

	defer file.Close()

	caCert, err := generateCACert(CertSubjectName)
	assert.NoError(err, "Failed to generate test certificate")

	_, err = file.Write(caCert)
	assert.NoError(err, "Failed to save certificate")

	recorder := &connectionRecorder{}
	options, err := aserto.NewConnectionOptions(aserto.WithCACertPath(caPath))
	assert.NoError(err)
	aserto.InternalNewConnection(recorder.Connect, options) //nolint: errcheck

	inPool, err := subjectInCertPool(recorder.tlsConf.RootCAs, CertSubjectName)
	if err != nil {
		t.Errorf("Failed to read root CAs: %s", err)
	}

	assert.True(inPool, "Aserto cert should be in root CAs")
}

func TestWithCACertPathAndInsecure(t *testing.T) {
	assert := assrt.New(t)
	tempdir := t.TempDir()
	caPath := fmt.Sprintf("%s/ca.pem", tempdir)

	file, err := os.OpenFile(caPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	assert.NoError(err, "Failed to create CA file")

	defer file.Close()

	caCert, err := generateCACert(CertSubjectName)
	assert.NoError(err, "Failed to generate test certificate")

	_, err = file.Write(caCert)
	assert.NoError(err, "Failed to save certificate")

	recorder := &connectionRecorder{}
	options, err := aserto.NewConnectionOptions(aserto.WithCACertPath(caPath), aserto.WithInsecure(true))
	assert.NoError(err)
	aserto.InternalNewConnection(recorder.Connect, options) //nolint: errcheck

	assert.Nil(recorder.tlsConf.RootCAs, "Aserto cert should be nil")
	assert.True(recorder.tlsConf.InsecureSkipVerify)
}

func TestWithDialOptions(t *testing.T) {
	assert := assrt.New(t)
	recorder := &connectionRecorder{}
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())

	options, err := aserto.NewConnectionOptions(aserto.WithDialOptions(creds))
	assert.NoError(err)
	aserto.InternalNewConnection(recorder.Connect, options) //nolint: errcheck
	assert.Contains(recorder.dialOptions, creds)
}

func generateCACert(subjectName string) ([]byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{subjectName},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	out := &bytes.Buffer{}
	if err := pem.Encode(out, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, fmt.Errorf("Failed to PEM encode certificate: %w", err)
	}

	return out.Bytes(), nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func subjectInCertPool(pool *x509.CertPool, name string) (bool, error) {
	for _, subject := range pool.Subjects() { //nolint: staticcheck
		var rdns pkix.RDNSequence

		_, err := asn1.Unmarshal(subject, &rdns)
		if err != nil {
			return false, err
		}

		for _, nameset := range rdns {
			for _, attr := range nameset {
				if attr.Value == name {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
