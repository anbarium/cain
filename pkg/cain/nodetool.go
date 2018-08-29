package cain

import (
	"fmt"
	"log"
	"strings"

	"github.com/maorfr/cain/pkg/utils"
	"github.com/maorfr/skbn/pkg/skbn"
	skbn_utils "github.com/maorfr/skbn/pkg/utils"
)

// TakeSnapshots takes a snapshot using nodetool in all pods in parallel
func TakeSnapshots(iClient interface{}, pods []string, namespace, container, keyspace string) string {
	k8sClient := iClient.(*skbn.K8sClient)
	tag := utils.GetTimeStamp()
	bwgSize := len(pods)
	bwg := skbn_utils.NewBoundedWaitGroup(bwgSize)
	for _, pod := range pods {
		bwg.Add(1)

		go func(k8sClient *skbn.K8sClient, namespace, pod, container, keyspace, tag string) {
			if err := takeSnapshot(k8sClient, namespace, pod, container, keyspace, tag); err != nil {
				log.Fatal(err)
			}
			bwg.Done()
		}(k8sClient, namespace, pod, container, keyspace, tag)
	}
	bwg.Wait()

	return tag
}

// ClearSnapshots clears a snapshot using nodetool in all pods in parallel
func ClearSnapshots(iClient interface{}, pods []string, namespace, container, keyspace, tag string) {
	k8sClient := iClient.(*skbn.K8sClient)
	bwgSize := len(pods)
	bwg := skbn_utils.NewBoundedWaitGroup(bwgSize)
	for _, pod := range pods {
		bwg.Add(1)

		go func(k8sClient *skbn.K8sClient, namespace, pod, container, keyspace, tag string) {
			if err := clearSnapshot(k8sClient, namespace, pod, container, keyspace, tag); err != nil {
				log.Fatal(err)
			}
			bwg.Done()
		}(k8sClient, namespace, pod, container, keyspace, tag)
	}
	bwg.Wait()
}

func RefreshTables(iClient interface{}, namespace, container, keyspace string, pods, tables []string) error {
	k8sClient := iClient.(*skbn.K8sClient)
	for _, pod := range pods {
		for _, table := range tables {
			if err := refreshTable(k8sClient, namespace, pod, container, keyspace, table); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetClusterName gets the name of the cassandra cluster
func GetClusterName(iClient interface{}, namespace, pod, container string) (string, error) {
	k8sClient := iClient.(*skbn.K8sClient)
	option := fmt.Sprintf("describecluster")
	output, err := nodetool(k8sClient, namespace, pod, container, option)
	if err != nil {
		return "", err
	}

	subStr := "Name:"
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, subStr) {
			output = strings.TrimSpace(strings.Replace(line, subStr, "", 1))
			break
		}
	}

	return output, nil
}

func takeSnapshot(k8sClient *skbn.K8sClient, namespace, pod, container, keyspace, tag string) error {
	log.Println(pod, "Taking snapshot")
	option := fmt.Sprintf("snapshot -t %s %s", tag, keyspace)
	output, err := nodetool(k8sClient, namespace, pod, container, option)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(output, "\n") {
		if line != "" {
			log.Println(pod, line)
		}
	}
	return nil
}

func clearSnapshot(k8sClient *skbn.K8sClient, namespace, pod, container, keyspace, tag string) error {
	log.Println(pod, "Clearing snapshot")
	option := fmt.Sprintf("clearsnapshot -t %s %s", tag, keyspace)
	output, err := nodetool(k8sClient, namespace, pod, container, option)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(output, "\n") {
		if line != "" {
			log.Println(pod, line)
		}
	}
	return nil
}

func refreshTable(k8sClient *skbn.K8sClient, namespace, pod, container, keyspace, table string) error {
	option := fmt.Sprintf("refresh %s %s", table, keyspace)
	output, err := nodetool(k8sClient, namespace, pod, container, option)
	if err != nil {
		return err
	}
	fmt.Println(pod, output)
	return nil
}

func nodetool(k8sClient *skbn.K8sClient, namespace, pod, container, option string) (string, error) {
	command := fmt.Sprintf("nodetool %s", option)
	stdout, stderr, err := skbn.Exec(*k8sClient, namespace, pod, container, command, nil)
	if len(stderr) != 0 {
		return "", fmt.Errorf("STDERR: " + (string)(stderr))
	}
	if err != nil {
		return "", err
	}

	return (string)(stdout), nil
}
