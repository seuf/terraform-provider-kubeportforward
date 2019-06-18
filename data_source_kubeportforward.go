package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func dataSourceKubePortForward() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceKubePortForwardRead,
		Schema: map[string]*schema.Schema{
			"kube_config": &schema.Schema{
				Type:        schema.TypeString,
				Default:     "",
				Optional:    true,
				Description: "kubectl config file. default to $HOME/.kube/config",
			},
			"namespace": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Namespace where the service to port forward is located",
			},
			"service": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service to port forward",
			},
			"local_port": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The local bind port (e.g. 3000). Default to randomly available port",
			},
			"remote_port": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The remote bind port (e.g. 3000)",
			},
			"port_forwarded": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceKubePortForwardRead(d *schema.ResourceData, meta interface{}) error {
	namespace := d.Get("namespace").(string)
	serviceName := d.Get("service").(string)
	localPort := d.Get("local_port").(string)
	remotePort := d.Get("remote_port").(string)
	portForwarded := d.Get("port_forwarded").(bool)

	log.Printf("[DEBUG] namespace: %v", namespace)
	log.Printf("[DEBUG] service: %v", serviceName)
	log.Printf("[DEBUG] localPort: %v", localPort)
	log.Printf("[DEBUG] remotePort: %v", remotePort)
	log.Printf("[DEBUG] portForwarded: %v", portForwarded)

	d.SetId(localPort)
	// Get Kubeconfig file
	var kubeconfig string
	kubeconfig = d.Get("kube_config").(string)
	if kubeconfig == "" {
		if home := homeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			return fmt.Errorf("[ERROR] No kubeconfig file specified and HOME environment variable not set")
		}
	}
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return fmt.Errorf("[ERROR] kubectl config file %s does not exists", kubeconfig)
	}
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("[ERROR] Can't Load kubectl config file %s : %s", kubeconfig, err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("Can't initalize kubernetes client : %s", err.Error())
	}

	if portForwarded == false {
		// Get k8s service
		log.Printf("Getting service %s in namespace %s", serviceName, namespace)
		svc, err := clientset.CoreV1().Services(namespace).Get(serviceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("Can't get kubernetes service %s in namespace %s : %s", serviceName, namespace, err.Error())
		}

		// Get k8s pods for service
		selector := mapToSelectorStr(svc.Spec.Selector)
		if selector == "" {
			return fmt.Errorf("ERROR: No backing pods for service %s in %s on cluster %s.\n", svc.Name, svc.Namespace, svc.ClusterName)
		}
		log.Printf("Getting pods with selector %s", selector)
		pods, err := clientset.CoreV1().Pods(svc.Namespace).List(metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return fmt.Errorf("ERROR: No pods found for %s: %s\n", selector, err.Error())
		}

		if len(pods.Items) < 1 {
			return fmt.Errorf("ERROR: No pods returned for service %s in %s on cluster %s.\n", svc.Name, svc.Namespace, svc.ClusterName)
		}

		podFound := false
		var podStatus v1.PodPhase
		for _, pod := range pods.Items {

			if pod.Status.Phase != v1.PodRunning {
				podStatus = pod.Status.Phase
				continue
			}
			podFound = true
			podPort := ""

			for _, port := range svc.Spec.Ports {

				podPort = port.TargetPort.String()

				if _, err := strconv.Atoi(podPort); err != nil {
					// search a pods containers for the named port
					if namedPodPort, ok := portSearch(podPort, pod.Spec.Containers); ok == true {
						podPort = namedPodPort
					}
				}

				if podPort != remotePort {
					continue
				}
				log.Printf("Found pod with remote port %s", pod.Name)
				// Create dialer
				path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, pod.Name)
				hostIP := strings.TrimLeft(config.Host, "https://")
				serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}
				transport, upgrader, err := spdy.RoundTripperFor(config)
				if err != nil {
					return fmt.Errorf("Can't create dialer : %s", err.Error())
				}
				dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", &serverURL)

				// Channels
				stopChannel := make(chan struct{}, 1)
				readyChannel := make(chan struct{})

				if localPort == "" {
					localPort = "0"
				}
				fwdPorts := []string{fmt.Sprintf("%s:%s", localPort, remotePort)}

				fw, err := portforward.New(dialer, fwdPorts, stopChannel, readyChannel, os.Stdout, os.Stderr)
				if err != nil {
					return fmt.Errorf("Can't init port foward :%s", err.Error())
				}

				signals := make(chan os.Signal, 1)
				signal.Notify(signals, os.Interrupt)
				defer signal.Stop(signals)

				go func() {
					<-signals
					if stopChannel != nil {
						log.Println("[KUBEPORTFORWARD] Closing Kube Port Forward Channel")
						fw.Close()
					}
				}()

				d.Set("port_forwarded", true)
				//wg.Add(1)

				go fw.ForwardPorts()

				time.Sleep(time.Second)
				log.Printf("Kube Port forwarded")
				//wg.Wait()
				// fwdedPorts, err := fw.GetPorts()
				// if err != nil {
				// 	return fmt.Errorf("Can't get forwarded ports %s", err.Error())
				// }
				// d.Set("local_port", fwdedPorts)
			}
		}
		if podFound == false {
			return fmt.Errorf("No ready pod Found %s", podStatus)
		}
	} else {
		return fmt.Errorf("Port already Forwarded")
	}

	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func mapToSelectorStr(msel map[string]string) string {
	selector := ""
	for k, v := range msel {
		if selector != "" {
			selector = selector + ","
		}
		selector = selector + fmt.Sprintf("%s=%s", k, v)
	}

	return selector
}

func portSearch(portName string, containers []v1.Container) (string, bool) {

	for _, container := range containers {
		for _, cp := range container.Ports {
			if cp.Name == portName {
				return fmt.Sprint(cp.ContainerPort), true
			}
		}
	}

	return "", false
}
