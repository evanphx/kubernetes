/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugins

import (
	"fmt"
	"strconv"

	"k8s.io/api/core/v1"
)

const (
	// AWSEBSDriverName is the name of the CSI driver for EBS
	AWSEBSDriverName = "ebs.csi.aws.com"
	// AWSEBSInTreePluginName is the name of the intree plugin for EBS
	AWSEBSInTreePluginName = "kubernetes.io/aws-ebs"
)

// AWSEBS handles translation of PV spec from In-tree EBS to CSI EBS and vice versa
type AWSEBS struct{}

// TranslateInTreePVToCSI takes a PV with AWSElasticBlockStore set from in-tree
// and converts the AWSElasticBlockStore source to a CSIPersistentVolumeSource
func (t *AWSEBS) TranslateInTreePVToCSI(pv *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	if pv == nil || pv.Spec.AWSElasticBlockStore == nil {
		return nil, fmt.Errorf("pv is nil or AWS EBS not defined on pv")
	}

	ebsSource := pv.Spec.AWSElasticBlockStore

	csiSource := &v1.CSIPersistentVolumeSource{
		Driver:       AWSEBSDriverName,
		VolumeHandle: ebsSource.VolumeID,
		ReadOnly:     ebsSource.ReadOnly,
		FSType:       ebsSource.FSType,
		VolumeAttributes: map[string]string{
			"partition": strconv.FormatInt(int64(ebsSource.Partition), 10),
		},
	}

	pv.Spec.AWSElasticBlockStore = nil
	pv.Spec.CSI = csiSource
	return pv, nil
}

// TranslateCSIPVToInTree takes a PV with CSIPersistentVolumeSource set and
// translates the EBS CSI source to a AWSElasticBlockStore source.
func (t *AWSEBS) TranslateCSIPVToInTree(pv *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	if pv == nil || pv.Spec.CSI == nil {
		return nil, fmt.Errorf("pv is nil or CSI source not defined on pv")
	}

	csiSource := pv.Spec.CSI

	ebsSource := &v1.AWSElasticBlockStoreVolumeSource{
		VolumeID: csiSource.VolumeHandle,
		FSType:   csiSource.FSType,
		ReadOnly: csiSource.ReadOnly,
	}

	if partition, ok := csiSource.VolumeAttributes["partition"]; ok {
		partValue, err := strconv.Atoi(partition)
		if err != nil {
			return nil, fmt.Errorf("Failed to convert partition %v to integer: %v", partition, err)
		}
		ebsSource.Partition = int32(partValue)
	}

	pv.Spec.CSI = nil
	pv.Spec.AWSElasticBlockStore = ebsSource
	return pv, nil
}

// CanSupport tests whether the plugin supports a given volume
// specification from the API.  The spec pointer should be considered
// const.
func (t *AWSEBS) CanSupport(pv *v1.PersistentVolume) bool {
	return pv != nil && pv.Spec.AWSElasticBlockStore != nil
}

// GetInTreePluginName returns the name of the intree plugin driver
func (t *AWSEBS) GetInTreePluginName() string {
	return AWSEBSInTreePluginName
}
