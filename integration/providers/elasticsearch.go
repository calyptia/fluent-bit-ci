package providers

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"strings"
	"time"
)

type ElasticSearchSuite struct {
	*BaseFluentbitSuite
}

const DefaultElasticsearchConfig =`
  service: |
    [SERVICE] 
        Flush        5
        Daemon       Off
        Log_Level    debug
        HTTP_Server On
        HTTP_Listen 0.0.0.0
        HTTP_Port {{ .Values.service.port }}
  inputs: |
    [INPUT]
        Name   dummy
        Tag    dummy.log
        Dummy  {"message":"testing"}
        Rate   10
  outputs: |
    [OUTPUT]
        Name    es
        Match   *
        Host    elasticsearch-master
        Port    9200
        Index   fluentbit
 `

const defaultRetries = 10
const defaultSleepPeriod = 30 * time.Second

func (suite *ElasticSearchSuite) TearDownTest() {
	suite.RemoveCharts()
}

func (suite *ElasticSearchSuite) SetupTest() {

	suite.InstallChart(ChartToInstall{fmt.Sprintf("elasticsearch-%s", strings.ToLower(random.UniqueId())), "https://helm.elastic.co","elastic", "elastic/elasticsearch",&helm.Options{
		KubectlOptions: suite.kubectlOpts,
		SetValues: map[string]string{"replicas": "1", "minMasterNodes": "1"}}})

	k8s.WaitUntilPodAvailable(suite.T(), suite.kubectlOpts, "elasticsearch-master-0", defaultRetries, defaultSleepPeriod)

	suite.InstallChart(ChartToInstall{fmt.Sprintf("fluent-bit-%s", strings.ToLower(random.UniqueId())),"https://fluent.github.io/helm-charts",
		"fluent","fluent/fluent-bit",suite.helmOpts	})

	k8s.WaitUntilPodAvailable(suite.T(), suite.kubectlOpts, suite.GetPodNameByChartRelease("fluent"), defaultRetries, defaultSleepPeriod)
}

const elasticSearchSleepPeriod = 15 * time.Second

func (suite *ElasticSearchSuite) TestFluentbitOutputToElasticSearch() {
	suite.assertHTTPResponseFromPod("/fluentbit/_search/", 9200, "elasticsearch-master-0", 200, defaultRetries, elasticSearchSleepPeriod)
}
