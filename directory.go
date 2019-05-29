package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// CollectObjectsFromDir scans all the files in a directory (including sub-directories), parse yaml|yml manifests
// and collect present objects and their names to map
func CollectObjectsFromDir(directories []string) (map[string][]string, error) {
	resourceMap := make(map[string][]string)

	for _, directory := range directories {

		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			var ext string
			if ext = filepath.Ext(path); ext != ".yaml" && ext != ".yml" {
				return nil
			}

			acceptedK8sTypes := regexp.MustCompile(acceptedK8sKinds)
			fileAsString, err := ioutil.ReadFile(path)
			if err != nil {
				log.Println(fmt.Sprintf("Error while reading YAML manifest. Err was: %s", err))
			}
			sepYamlfiles := strings.Split(string(fileAsString), "---")

			for _, file := range sepYamlfiles {
				if file == "\n" || file == "" {
					// ignore empty cases
					continue
				}

				decode := scheme.Codecs.UniversalDeserializer().Decode
				obj, groupVersionKind, err := decode([]byte(file), nil, nil)
				if err != nil {
					log.Println(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
					continue
				}

				if !acceptedK8sTypes.MatchString(groupVersionKind.Kind) {
					log.Printf("Skipping object with type: %s", groupVersionKind.Kind)
				} else {

					switch obj.(type) {
					case *appsv1.Deployment:
						resourceMap[groupVersionKind.Kind] = append(resourceMap[groupVersionKind.Kind], obj.(*appsv1.Deployment).ObjectMeta.Name)
					case *corev1.Service:
						resourceMap[groupVersionKind.Kind] = append(resourceMap[groupVersionKind.Kind], obj.(*corev1.Service).ObjectMeta.Name)
					case *appsv1.StatefulSet:
						resourceMap[groupVersionKind.Kind] = append(resourceMap[groupVersionKind.Kind], obj.(*appsv1.StatefulSet).ObjectMeta.Name)
					case *v1beta1.CronJob:
						resourceMap[groupVersionKind.Kind] = append(resourceMap[groupVersionKind.Kind], obj.(*v1beta1.CronJob).ObjectMeta.Name)
					case *corev1.LimitRange:
						resourceMap[groupVersionKind.Kind] = append(resourceMap[groupVersionKind.Kind], obj.(*corev1.LimitRange).ObjectMeta.Name)
					case *appsv1.DaemonSet:
						resourceMap[groupVersionKind.Kind] = append(resourceMap[groupVersionKind.Kind], obj.(*appsv1.DaemonSet).ObjectMeta.Name)
					default:
						log.Printf("Skip type: %s", groupVersionKind.Kind)
					}
				}
			}

			return nil
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	return resourceMap, nil
}

// Except returns a new slice, containing all items that are in the left slice (k8scluster) but not the right slice (VCS)
func Except(left, right []string) []string {
	for i := len(left) - 1; i >= 0; i-- {
		for _, vD := range right {
			if left[i] == vD {
				left = append(left[:i], left[i+1:]...)
				break
			}
		}
	}
	return left
}
