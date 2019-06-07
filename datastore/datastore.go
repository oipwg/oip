package datastore

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/azer/logger"
	"github.com/gobuffalo/packr/v2"
	"github.com/oipwg/oip/config"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/olivere/elastic.v6"
)

var client *elastic.Client
var AutoBulk BulkIndexer

var mappings = make(map[string]string)
var mapBox = packr.New("mappings", "./mappings")

func Setup(ctx context.Context) error {
	var err error

	httpClient, err := getHttpClient()
	if err != nil {
		log.Error("couldn't create httpClient", logger.Attrs{"err": err})
		return err
	}

	client, err = elastic.NewClient(elastic.SetSniff(false), elastic.SetHttpClient(httpClient),
		elastic.SetURL(viper.GetString("elastic.host")))
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
	useCert := viper.GetBool("elastic.useCert")
	if useCert {
		certFile := config.GetFilePath("elastic.certFile")
		certKey := config.GetFilePath("elastic.certKey")
		rootCertPath := config.GetFilePath("elastic.certRoot")

		// ToDo: add encrypted key support - potentially via x509.DecryptPEMBloc & tls.ParsePKCS1PrivateKey
		cert, err := tls.LoadX509KeyPair(certFile, certKey)
		if err != nil {
			log.Error("couldn't LoadX509KeyPair", logger.Attrs{"err": err})
			return nil, err
		}
		caCert, err := ioutil.ReadFile(rootCertPath)
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

func RegisterMapping(index, fileName string) {
	index = Index(index) // apply proper prefix
	mapping, err := mapBox.FindString(fileName)
	if err != nil {
		panic(fmt.Sprintf("Unable to find mapping %s for index %s", fileName, index))
	}
	mappings[index] = mapping
	if client != nil {
		err := createIndex(context.TODO(), index, mapping)
		if err != nil {
			panic(fmt.Sprintf("unable to create index %s - %s", index, err))
		}
	}
}

func createIndex(ctx context.Context, index, mapping string) error {
	exists, err := client.IndexExists(Index(index)).Do(ctx)
	if err != nil {
		return errors.Wrap(err, "index existence check failure")
	}

	if !exists {
		createIndex, err := client.CreateIndex(Index(index)).BodyString(mapping).Do(ctx)
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

func Index(index string) string {
	if config.IsTestnet() {
		if strings.HasPrefix(index, "testnet-") {
			return index
		}
		return "testnet-" + index
	}
	if strings.HasPrefix(index, "mainnet-") {
		return index
	}
	return "mainnet-" + index
}
