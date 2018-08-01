package datastore

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/azer/logger"
	"github.com/bitspill/oip/config"
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v6"
)

var client *elastic.Client
var AutoBulk BulkIndexer

var mappings = make(map[string]string)

func Setup(ctx context.Context) error {
	var err error

	httpClient, err := getHttpClient()
	if err != nil {
		log.Error("couldn't create httpClient", logger.Attrs{"err": err})
		return err
	}

	client, err = elastic.NewClient(elastic.SetSniff(false), elastic.SetHttpClient(httpClient),
		elastic.SetURL(config.Elastic.Host))
	if err != nil {
		log.Error("unable to connect to elasticsearch", logger.Attrs{"err": err})
		return errors.Wrap(err, "datastore.setup.newClient")
	}

	for index, mapping := range mappings {
		err := createIndex(ctx, index, mapping)
		if err != nil {
			return errors.Wrap(err, fmt.Sprint("datastore.setup.createIndex", index))
		}
	}

	AutoBulk = BeginBulkIndexer()

	return nil
}

func getHttpClient() (*http.Client, error) {
	var httpClient *http.Client
	if config.Elastic.UseCert {
		// ToDo: add encrypted key support - potentially via x509.DecryptPEMBloc & tls.ParsePKCS1PrivateKey
		cert, err := tls.LoadX509KeyPair(config.Elastic.CertFile, config.Elastic.CertKey)
		if err != nil {
			log.Error("couldn't LoadX509KeyPair", logger.Attrs{"err": err})
			return nil, err
		}
		caCert, err := ioutil.ReadFile(config.Elastic.CertRoot)
		if err != nil {
			log.Error("couldn't read root certificate", logger.Attrs{"err": err})
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: true,
		}
		tlsConfig.BuildNameToCertificate()
		transport := &http.Transport{
			TLSClientConfig: tlsConfig,
		}

		httpClient = &http.Client{
			Transport: transport,
		}
	} else {
		httpClient = http.DefaultClient
	}

	return httpClient, nil
}

func RegisterMapping(index string, mapping string) error {
	mappings[index] = mapping
	if client != nil {
		return createIndex(context.TODO(), index, mapping)
	}
	return nil
}

func createIndex(ctx context.Context, index string, mapping string) error {
	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		return errors.Wrap(err, "index existence check failure")
	}

	if !exists {
		createIndex, err := client.CreateIndex(index).BodyString(mapping).Do(ctx)
		if err != nil {
			return errors.Wrap(err, "create index failed")
		}
		if !createIndex.Acknowledged {
			return errors.New("create index not acknowledged")
		}
	}

	return nil
}

func Client() *elastic.Client {
	return client
}
