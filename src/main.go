package main

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"flag"
	"log"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	// SynchronizationScheduleSeconds The time in seconds between synchronizations
	SynchronizationScheduleSeconds = 20
)

// arrayFlags type allows to call a '--flag' more than once
type arrayFlags []string

// String returns a string representation of the type for the 'flag' library
func (i *arrayFlags) String() string {
	result := strings.Join(*i, " ")
	return result
}

// Set defines how an element of the type must be treated when is being set by the 'flag' library
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var SecretNames arrayFlags
var TLSHosts arrayFlags

// Secret resource structure that will be created or updated in Kubernetes
type Secret struct {
	Name        string
	Namespace   string `json:"namespace,omitempty"`
	Certificate string `json:"tls.crt"`
}

// BuildSecrets Get the TLS certificates from hosts on TLSHostsAddresses
// and craft the Secret Kubernetes resources with them
func BuildSecrets(SecretNames []string, TLSHostsAddresses []string) ([]Secret, error) {
	var secrets []Secret

	// Establish TLS connection
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	// Get the certificates from hosts to fill the Secrets
	for item, TLSHost := range TLSHostsAddresses {
		conn, err := tls.Dial("tcp", TLSHost, conf)
		if err != nil {
			return secrets, err
		}

		defer conn.Close()

		certs := conn.ConnectionState().PeerCertificates
		if len(certs) == 0 {
			return secrets, err
		}

		pemContent := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certs[0].Raw,
		})
		secrets = append(secrets, Secret{Name: SecretNames[item], Certificate: string(pemContent)})
	}

	return secrets, nil
}

// GetKubernetesClient Return a Kubernetes client configured to connect from inside or outside the cluster
func GetKubernetesClient(connectionMode string, kubeconfigPath string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var client *kubernetes.Clientset

	// Create configuration to connect from inside the cluster using Kubernetes mechanisms
	config, err := rest.InClusterConfig()

	// Create configuration to connect from outside the cluster, using kubectl
	if connectionMode == "kubectl" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	// Check configuration errors in both cases
	if err != nil {
		return client, err
	}

	// Construct the client
	client, err = kubernetes.NewForConfig(config)
	return client, err
}

func SynchronizeSecrets(client *kubernetes.Clientset, namespace string, secrets []Secret) error {
	var err error
	secretsClient := client.CoreV1().Secrets(namespace)

	// Synchronize the Secret resources
	for _, secret := range secrets {

		// Generate a Secret structure
		secretObject := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secret.Name,
				Namespace: namespace,
			},
			StringData: map[string]string{
				"tls.crt": secret.Certificate,
			},
		}

		// Search for the secret
		_, err = secretsClient.Get(context.Background(), secret.Name, metav1.GetOptions{})

		// The Secret does NOT exist: Create it
		if err != nil {
			_, err = secretsClient.Create(context.Background(), secretObject.DeepCopy(), metav1.CreateOptions{})
			if err != nil {
				return err
			}
		}

		// The Secret DOES exist: Update it
		_, err = secretsClient.Update(context.Background(), secretObject, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return err
}

func main() {
	// Get the values from flags
	connectionMode := flag.String("connection-mode", "kubectl", "(optional) What type of connection to use: incluster, kubectl")
	kubeconfig := flag.String("kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	namespaceFlag := flag.String("namespace", "default", "Kubernetes Namespace where to synchronize the certificates")
	flag.Var(&SecretNames, "secret-name", "Name of the Kubernetes Secret that will be created with the PEM information")
	flag.Var(&TLSHosts, "tls-host", "HOST:PORT that will be dialed to get the TLS certificate")
	flag.Parse()

	// Force to have a name per each TLS host
	if len(SecretNames) != len(TLSHosts) {
		log.Println("The number of Secrets to generate and the TLS hosts to visit must be the same")
		return
	}

	// Generate the Kubernetes client to modify the resources
	log.Printf("Generating the client to connect to Kubernetes")
	client, err := GetKubernetesClient(*connectionMode, *kubeconfig)
	if err != nil {
		log.Printf("Error connecting to Kubernetes API: %s", err)
	}

	// Update the Secrets time by time
	for {
		// Build the Secret resources with the certificates content
		secrets, err := BuildSecrets(SecretNames, TLSHosts)

		// Use the Kubernetes client to synchronize the resources
		log.Printf("Synchronizing the Secrets in the namespace: %s", *namespaceFlag)
		err = SynchronizeSecrets(client, *namespaceFlag, secrets)
		if err != nil {
			log.Printf("Error synchronizing the Secrets: %s", err)
		}

		log.Printf("Next synchronization in %d seconds", SynchronizationScheduleSeconds)
		time.Sleep(SynchronizationScheduleSeconds * time.Second)
	}
}
