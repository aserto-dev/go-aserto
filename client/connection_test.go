package client_test

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

	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/header"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type dialRecorder struct {
	context     context.Context
	address     string
	tlsConf     *tls.Config
	callerCreds credentials.PerRPCCredentials
	connection  *client.Connection
	dialOptions []grpc.DialOption
}

func (d *dialRecorder) DialContext(
	ctx context.Context,
	address string,
	tlsConf *tls.Config,
	callerCreds credentials.PerRPCCredentials,
	connection *client.Connection,
	options []grpc.DialOption,
) (*grpc.ClientConn, error) {
	d.context = ctx
	d.address = address
	d.tlsConf = tlsConf
	d.callerCreds = callerCreds
	d.connection = connection
	d.dialOptions = options

	return &grpc.ClientConn{}, nil
}

func TestWithAddr(t *testing.T) {
	recorder := &dialRecorder{}
	options, err := client.NewConnectionOptions(client.WithAddr("address"))
	assert.NoError(t, err)

	client.InternalNewConnection(context.TODO(), recorder.DialContext, options) //nolint: errcheck

	assert.Equal(t, "address", recorder.address)
}

func TestWithURL(t *testing.T) {
	recorder := &dialRecorder{}

	const URL = "https://server.com:123"
	svcURL, err := url.Parse(URL)
	assert.NoError(t, err)

	options, err := client.NewConnectionOptions(client.WithURL(svcURL))
	assert.NoError(t, err)
	client.InternalNewConnection(context.TODO(), recorder.DialContext, options) //nolint: errcheck

	assert.Equal(t, URL, recorder.address)
}

func TestAddrAndURL(t *testing.T) {
	svcURL, err := url.Parse("https://server.com:123")
	assert.NoError(t, err)

	_, err = client.NewConnectionOptions(client.WithAddr("address"), client.WithURL(svcURL))
	assert.Error(t, err)
}

func TestWithInsecure(t *testing.T) {
	recorder := &dialRecorder{}
	options, err := client.NewConnectionOptions(client.WithInsecure(true))
	assert.NoError(t, err)
	client.InternalNewConnection(context.TODO(), recorder.DialContext, options) //nolint: errcheck

	assert.True(t, recorder.tlsConf.InsecureSkipVerify)
}

func TestWithTokenAuth(t *testing.T) {
	recorder := &dialRecorder{}
	options, err := client.NewConnectionOptions(client.WithTokenAuth("<token>"))
	assert.NoError(t, err)
	client.InternalNewConnection(context.TODO(), recorder.DialContext, options) //nolint: errcheck

	md, err := recorder.callerCreds.GetRequestMetadata(context.TODO())
	assert.NoError(t, err)

	token, ok := md["authorization"]
	assert.True(t, ok)
	assert.Equal(t, "bearer <token>", token)
}

func TestWithBearerTokenAuth(t *testing.T) {
	recorder := &dialRecorder{}
	options, err := client.NewConnectionOptions(client.WithTokenAuth("bearer <token>"))
	assert.NoError(t, err)
	client.InternalNewConnection(context.TODO(), recorder.DialContext, options) //nolint: errcheck

	md, err := recorder.callerCreds.GetRequestMetadata(context.TODO())
	assert.NoError(t, err)

	token, ok := md["authorization"]
	assert.True(t, ok)
	assert.Equal(t, "bearer <token>", token)
}

func TestWithAPIKey(t *testing.T) {
	recorder := &dialRecorder{}
	options, err := client.NewConnectionOptions(client.WithAPIKeyAuth("<apikey>"))
	assert.NoError(t, err)
	client.InternalNewConnection(context.TODO(), recorder.DialContext, options) //nolint: errcheck

	md, err := recorder.callerCreds.GetRequestMetadata(context.TODO())
	assert.NoError(t, err)

	token, ok := md["authorization"]
	assert.True(t, ok)
	assert.Equal(t, "basic <apikey>", token)
}

func TestTokenAndAPIKey(t *testing.T) {
	_, err := client.NewConnectionOptions(client.WithAPIKeyAuth("<apikey>"), client.WithTokenAuth("<token>"))
	assert.Error(t, err)
}

func TestWithTenantID(t *testing.T) {
	recorder := &dialRecorder{}
	options, err := client.NewConnectionOptions(client.WithTenantID("<tenantid>"))
	assert.NoError(t, err)
	client.InternalNewConnection(context.TODO(), recorder.DialContext, options) //nolint: errcheck

	assert.Equal(t, "<tenantid>", recorder.connection.TenantID)

	ctx := context.TODO()
	recorder.connection.InternalUnary( //nolint: errcheck, dupl
		ctx,
		"method",
		"request",
		"reply",
		recorder.connection.Conn,
		func(c context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(c)
			assert.True(t, ok)

			tenantID := md.Get("aserto-tenant-id")
			assert.Equal(t, 1, len(tenantID), "request should contain tenant ID metadata field")
			assert.Equal(t, "<tenantid>", tenantID[0], "tenant ID metadata should have the expected value")

			assert.Equal(t, "method", method, "'method' parameter should be a passthrough")
			assert.Equal(t, "request", req, "'request' parameter should be a passthrough")
			assert.Equal(t, "reply", reply, "'reply' parameter should be a passthrough")
			assert.Equal(t, recorder.connection.Conn, cc)

			return nil
		})

	recorder.connection.InternalStream( //nolint: errcheck
		ctx,
		nil,
		recorder.connection.Conn,
		"method",
		func(
			c context.Context,
			desc *grpc.StreamDesc,
			cc *grpc.ClientConn,
			method string,
			opts ...grpc.CallOption,
		) (grpc.ClientStream, error) {
			md, ok := metadata.FromOutgoingContext(c)
			assert.True(t, ok)

			tenantID := md.Get("aserto-tenant-id")
			assert.Equal(t, 1, len(tenantID), "request should contain tenant ID metadata field")
			assert.Equal(t, "<tenantid>", tenantID[0], "tenant ID metadata should have the expected value")

			assert.Equal(t, "method", method, "'method' parameter should be a passthrough")
			assert.Equal(t, recorder.connection.Conn, cc)

			return nil, nil
		},
	)
}

func TestWithSessionID(t *testing.T) {
	recorder := &dialRecorder{}
	options, err := client.NewConnectionOptions(client.WithSessionID("<sessionid>"))
	assert.NoError(t, err)
	client.InternalNewConnection(context.TODO(), recorder.DialContext, options) //nolint: errcheck

	assert.Equal(t, "<sessionid>", recorder.connection.SessionID)

	ctx := context.TODO()
	recorder.connection.InternalUnary( //nolint: errcheck, dupl
		ctx,
		"method",
		"request",
		"reply",
		recorder.connection.Conn,
		func(c context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(c)
			assert.True(t, ok)

			sessionID := md.Get("aserto-session-id")
			assert.Equal(t, 1, len(sessionID), "request should contain session ID metadata field")
			assert.Equal(t, "<sessionid>", sessionID[0], "session ID metadata should have the expected value")

			assert.Equal(t, "method", method, "'method' parameter should be a passthrough")
			assert.Equal(t, "request", req, "'request' parameter should be a passthrough")
			assert.Equal(t, "reply", reply, "'reply' parameter should be a passthrough")
			assert.Equal(t, recorder.connection.Conn, cc)

			return nil
		})

	recorder.connection.InternalStream( //nolint: errcheck
		ctx,
		nil,
		recorder.connection.Conn,
		"method",
		func(
			c context.Context,
			desc *grpc.StreamDesc,
			cc *grpc.ClientConn,
			method string,
			opts ...grpc.CallOption,
		) (grpc.ClientStream, error) {
			md, ok := metadata.FromOutgoingContext(c)
			assert.True(t, ok)

			sessionID := md.Get(string(header.HeaderAsertoSessionID))
			assert.Equal(t, 1, len(sessionID), "request should contain session ID metadata field")
			assert.Equal(t, "<sessionid>", sessionID[0], "session ID metadata should have the expected value")

			assert.Equal(t, "method", method, "'method' parameter should be a passthrough")
			assert.Equal(t, recorder.connection.Conn, cc)

			return nil, nil
		},
	)
}

const CertSubjectName = "Testing Inc."

func TestWithCACertPath(t *testing.T) {
	tempdir := t.TempDir()
	caPath := fmt.Sprintf("%s/ca.pem", tempdir)

	file, err := os.OpenFile(caPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	assert.NoError(t, err, "Failed to create CA file")

	defer file.Close()

	caCert, err := generateCACert(CertSubjectName)
	assert.NoError(t, err, "Failed to generate test certificate")

	_, err = file.Write(caCert)
	assert.NoError(t, err, "Failed to save certificate")

	recorder := &dialRecorder{}
	options, err := client.NewConnectionOptions(client.WithCACertPath(caPath))
	assert.NoError(t, err)
	client.InternalNewConnection(context.TODO(), recorder.DialContext, options) //nolint: errcheck

	inPool, err := subjectInCertPool(recorder.tlsConf.RootCAs, CertSubjectName)
	if err != nil {
		t.Errorf("Failed to read root CAs: %s", err)
	}

	assert.True(t, inPool, "Aserto cert should be in root CAs")
}

func TestWithCACertPathAndInsecure(t *testing.T) {
	tempdir := t.TempDir()
	caPath := fmt.Sprintf("%s/ca.pem", tempdir)

	file, err := os.OpenFile(caPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	assert.NoError(t, err, "Failed to create CA file")

	defer file.Close()

	caCert, err := generateCACert(CertSubjectName)
	assert.NoError(t, err, "Failed to generate test certificate")

	_, err = file.Write(caCert)
	assert.NoError(t, err, "Failed to save certificate")

	recorder := &dialRecorder{}
	options, err := client.NewConnectionOptions(client.WithCACertPath(caPath), client.WithInsecure(true))
	assert.NoError(t, err)
	client.InternalNewConnection(context.TODO(), recorder.DialContext, options) //nolint: errcheck

	assert.Nil(t, recorder.tlsConf.RootCAs, "Aserto cert should be nil")
	assert.True(t, recorder.tlsConf.InsecureSkipVerify)
}

func TestWithDialOptions(t *testing.T) {
	recorder := &dialRecorder{}
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())

	options, err := client.NewConnectionOptions(client.WithDialOptions(creds))
	assert.NoError(t, err)
	client.InternalNewConnection(context.TODO(), recorder.DialContext, options) //nolint: errcheck
	assert.Contains(t, recorder.dialOptions, creds)
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
