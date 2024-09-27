/*
Copyright (C) 2022-2024 ApeCloud Co., Ltd

This file is part of KubeBlocks project

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package controllerutil

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	"github.com/apecloud/kubeblocks/pkg/constant"
	"github.com/apecloud/kubeblocks/pkg/generics"
	testapps "github.com/apecloud/kubeblocks/pkg/testutil/apps"
)

var _ = Describe("cluster shard component", func() {

	// Cleanups
	cleanEnv := func() {
		// must wait till resources deleted and no longer existed before the testcases start,
		// otherwise if later it needs to create some new resource objects with the same name,
		// in race conditions, it will find the existence of old objects, resulting failure to
		// create the new objects.
		By("clean resources")

		// delete cluster(and all dependent sub-resources), cluster definition
		testapps.ClearClusterResourcesWithRemoveFinalizerOption(&testCtx)

		// delete rest mocked objects
		inNS := client.InNamespace(testCtx.DefaultNamespace)
		ml := client.HasLabels{testCtx.TestObjLabelKey}
		// namespaced
		testapps.ClearResourcesWithRemoveFinalizerOption(&testCtx, generics.ComponentSignature, true, inNS, ml)
		testapps.ClearResourcesWithRemoveFinalizerOption(&testCtx, generics.PodSignature, true, inNS, ml)
		testapps.ClearResourcesWithRemoveFinalizerOption(&testCtx, generics.ServiceSignature, true, inNS)
	}

	Context("cluster shard component", func() {
		const (
			compDefName           = "test-compdef"
			clusterName           = "test-cluster"
			mysqlCompName         = "mysql"
			mysqlShardingName     = "mysql-sharding"
			mysqlShardingCompName = "mysql-sharding-comp"
		)

		var (
			cluster *appsv1.Cluster
		)

		BeforeEach(func() {
			cleanEnv()

			testapps.NewComponentDefinitionFactory(compDefName).SetDefaultSpec().GetObject()
			cluster = testapps.NewClusterFactory(testCtx.DefaultNamespace, clusterName, "").
				SetUID(clusterName).
				AddComponent(mysqlCompName, compDefName).
				AddShardingSpec(mysqlShardingName, compDefName).
				SetShards(1).
				Create(&testCtx).GetObject()
		})

		It("generate sharding component spec test", func() {
			By("create mock sharding component object")
			mockCompObj := testapps.NewComponentFactory(testCtx.DefaultNamespace, cluster.Name+"-"+mysqlShardingCompName, "").
				AddAnnotations(constant.KBAppClusterUIDKey, string(cluster.UID)).
				AddLabels(constant.AppInstanceLabelKey, cluster.Name).
				AddLabels(constant.KBAppShardingNameLabelKey, mysqlShardingName).
				SetReplicas(1).
				Create(&testCtx).
				GetObject()
			compKey := client.ObjectKeyFromObject(mockCompObj)
			Eventually(testapps.CheckObjExists(&testCtx, compKey, &appsv1.Component{}, true)).Should(Succeed())

			shardingSpec := &appsv1.ShardingSpec{
				Template: appsv1.ClusterComponentSpec{
					Replicas: 2,
				},
				Name:   mysqlShardingName,
				Shards: 2,
			}
			shardingCompSpecList, err := GenShardingCompSpecList(testCtx.Ctx, k8sClient, cluster, shardingSpec)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(shardingCompSpecList).ShouldNot(BeNil())
			Expect(len(shardingCompSpecList)).Should(BeEquivalentTo(2))
		})
	})
})
