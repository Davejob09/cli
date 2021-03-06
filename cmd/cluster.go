/*
Copyright © 2020 CAST AI

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

package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/castai/cli/pkg/client"
	"github.com/castai/cli/pkg/client/sdk"
)

func newClusterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cluster",
		Short: "Manage clusters",
	}
}

func getClusterIDFromArgs(cmd *cobra.Command, api client.Interface) (string, error) {
	ctx := cmd.Context()
	if len(cmd.Flags().Args()) == 0 {
		cluster, err := selectCluster(ctx, api)
		if err != nil {
			return "", err
		}
		return cluster.Id, nil
	}

	value := cmd.Flags().Args()[0]
	return getClusterID(ctx, api, value)
}

func getClusterIDFromFlag(cmd *cobra.Command, api client.Interface) (string, error) {
	ctx := cmd.Context()
	value, err := cmd.Flags().GetString(flagCluster)
	if err != nil {
		return "", err
	}

	if value == "" {
		cluster, err := selectCluster(ctx, api)
		if err != nil {
			return "", err
		}
		return cluster.Id, nil
	}

	return getClusterID(ctx, api, value)
}

func getClusterID(ctx context.Context, api client.Interface, value string) (string, error) {
	uuidID, err := uuid.Parse(value)
	if err == nil {
		return uuidID.String(), nil
	}
	clusters, err := api.ListKubernetesClusters(ctx, &sdk.ListKubernetesClustersParams{})
	if err != nil {
		return "", err
	}
	for _, cluster := range clusters {
		if strings.ToLower(cluster.Name) == strings.ToLower(value) {
			return cluster.Id, nil
		}
	}
	return "", fmt.Errorf("clusterID for %s not found", value)
}

func selectCluster(ctx context.Context, api client.Interface) (*sdk.KubernetesCluster, error) {
	items, err := api.ListKubernetesClusters(ctx, &sdk.ListKubernetesClustersParams{})
	if err != nil {
		return nil, err
	}
	selectList := make([]string, len(items))
	for i, item := range items {
		selectList[i] = item.Name
	}

	var selected string
	prompt := &survey.Select{
		Message: "Select cluster:",
		Options: selectList,
		Default: selectList[0],
	}

	if err := survey.AskOne(prompt, &selected, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.Name == selected {
			return &item, nil
		}
	}

	return nil, errors.New("cluster not found")
}
