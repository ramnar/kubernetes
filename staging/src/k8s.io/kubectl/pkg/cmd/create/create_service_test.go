/*
Copyright 2016 The Kubernetes Authors.

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

package create

import (
	"net/http"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest/fake"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"
	"k8s.io/kubectl/pkg/scheme"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestCreateService(t *testing.T) {
	//ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
	tests := []struct {
		options  *ServiceClusterIPOpts
		expected *v1.Service
		expectErr   bool
	}{
		{
			options: &ServiceClusterIPOpts{
				ServiceCommon{
	            Name:        "clusterip-ok",
				TCP:         []string{"456", "321:908"},
				ClusterIP:   "",
				Type: v1.ServiceTypeClusterIP,
				},
				//ioStreams,
				//genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
			},
			expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "clusterip-ok",
					Labels: map[string]string{"app": "clusterip-ok"},
				},
				Spec: v1.ServiceSpec{Type: "ClusterIP",
					Ports: []v1.ServicePort{{Name: "456", Protocol: "TCP", Port: 456, TargetPort: intstr.IntOrString{Type: 0, IntVal: 456, StrVal: ""}, NodePort: 0},
						{Name: "321-908", Protocol: "TCP", Port: 321, TargetPort: intstr.IntOrString{Type: 0, IntVal: 908, StrVal: ""}, NodePort: 0}},
					Selector:  map[string]string{"app": "clusterip-ok"},
					ClusterIP: "", ExternalIPs: []string(nil), LoadBalancerIP: ""},
			},
			expectErr: false,
		},
		{
			options: &ServiceClusterIPOpts{
				ServiceCommon{
			  Name:        "clusterip-missing",
			  Type: v1.ServiceTypeClusterIP,
				},
				//				ioStreams,
				//genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
			},
			expectErr:   true,
		},
		{
			options: &ServiceClusterIPOpts{
				
			ServiceCommon{
			  Name:        "clusterip-none-wrong-type",
			  TCP:         []string{},
			  ClusterIP:   "None",
			  Type: v1.ServiceTypeNodePort,
			},		
				//	ioStreams,
				//genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
			},
			expectErr:   true,
		},		
		{
			options: &ServiceClusterIPOpts{
	            ServiceCommon{
	            Name:        "clusterip-none-ok",
				TCP:         []string{},
				ClusterIP:   "None",
				Type: v1.ServiceTypeClusterIP,
	            },
	            //				ioStreams,
				//genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
			},
			
			expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "clusterip-none-ok",
					Labels: map[string]string{"app": "clusterip-none-ok"},
				},
				Spec: v1.ServiceSpec{Type: "ClusterIP",
					Ports:     []v1.ServicePort{},
					Selector:  map[string]string{"app": "clusterip-none-ok"},
					ClusterIP: "None", ExternalIPs: []string(nil), LoadBalancerIP: ""},
			},
			expectErr: false,
		},
		{
			options: &ServiceClusterIPOpts{
	            ServiceCommon{
	            Name:        "clusterip-none-and-port-mapping",
				TCP:         []string{"456:9898"},
				ClusterIP:   "None",
				Type: v1.ServiceTypeClusterIP,
	            },
	            //				ioStreams,
				//genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
			},
			expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "clusterip-none-and-port-mapping",
					Labels: map[string]string{"app": "clusterip-none-and-port-mapping"},
				},
				Spec: v1.ServiceSpec{Type: "ClusterIP",
					Ports:     []v1.ServicePort{{Name: "456-9898", Protocol: "TCP", Port: 456, TargetPort: intstr.IntOrString{Type: 0, IntVal: 9898, StrVal: ""}, NodePort: 0}},
					Selector:  map[string]string{"app": "clusterip-none-and-port-mapping"},
					ClusterIP: "None", ExternalIPs: []string(nil), LoadBalancerIP: ""},
			},
			expectErr: false,
		},		
		
	}

	for _, tc := range tests {
		service, err := tc.options.createService()
		if err != nil {
			t.Errorf("unexpected error:\n%#v\n", err)
			return
		}
		if tc.expectErr && err == nil {
		   continue
	    }
		if !apiequality.Semantic.DeepEqual(service, tc.expected) {
			t.Errorf("expected:\n%#v\ngot:\n%#v", tc.expected, service)
		}
	}
}

//
//func TestCreateService(t *testing.T) {
//	service := &v1.Service{}
//	service.Name = "my-service"
//	tf := cmdtesting.NewTestFactory().WithNamespace("test")
//	defer tf.Cleanup()
//
//	codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
//	negSer := scheme.Codecs
//
//	tf.Client = &fake.RESTClient{
//		GroupVersion:         schema.GroupVersion{Version: "v1"},
//		NegotiatedSerializer: negSer,
//		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
//			switch p, m := req.URL.Path, req.Method; {
//			case p == "/namespaces/test/services" && m == "POST":
//				return &http.Response{StatusCode: http.StatusCreated, Header: cmdtesting.DefaultHeader(), Body: cmdtesting.ObjBody(codec, service)}, nil
//			default:
//				t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
//				return nil, nil
//			}
//		}),
//	}
//	ioStreams, _, buf, _ := genericclioptions.NewTestIOStreams()
//	cmd := NewCmdCreateServiceClusterIP(tf, ioStreams)
//	cmd.Flags().Set("output", "name")
//	cmd.Flags().Set("tcp", "8080:8000")
//	cmd.Run(cmd, []string{service.Name})
//	expectedOutput := "service/" + service.Name + "\n"
//	if buf.String() != expectedOutput {
//		t.Errorf("expected output: %s, but got: %s", expectedOutput, buf.String())
//	}
//}

func TestCreateServiceNodePort(t *testing.T) {
	service := &v1.Service{}
	service.Name = "my-node-port-service"
	tf := cmdtesting.NewTestFactory().WithNamespace("test")
	defer tf.Cleanup()

	codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
	negSer := scheme.Codecs

	tf.Client = &fake.RESTClient{
		GroupVersion:         schema.GroupVersion{Version: "v1"},
		NegotiatedSerializer: negSer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch p, m := req.URL.Path, req.Method; {
			case p == "/namespaces/test/services" && m == http.MethodPost:
				return &http.Response{StatusCode: http.StatusCreated, Header: cmdtesting.DefaultHeader(), Body: cmdtesting.ObjBody(codec, service)}, nil
			default:
				t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
				return nil, nil
			}
		}),
	}
	ioStreams, _, buf, _ := genericclioptions.NewTestIOStreams()
	cmd := NewCmdCreateServiceNodePort(tf, ioStreams)
	cmd.Flags().Set("output", "name")
	cmd.Flags().Set("tcp", "30000:8000")
	cmd.Run(cmd, []string{service.Name})
	expectedOutput := "service/" + service.Name + "\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output: %s, but got: %s", expectedOutput, buf.String())
	}
}

func TestCreateServiceExternalName(t *testing.T) {
	service := &v1.Service{}
	service.Name = "my-external-name-service"
	tf := cmdtesting.NewTestFactory().WithNamespace("test")
	defer tf.Cleanup()

	codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
	negSer := scheme.Codecs

	tf.Client = &fake.RESTClient{
		GroupVersion:         schema.GroupVersion{Version: "v1"},
		NegotiatedSerializer: negSer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch p, m := req.URL.Path, req.Method; {
			case p == "/namespaces/test/services" && m == http.MethodPost:
				return &http.Response{StatusCode: http.StatusCreated, Header: cmdtesting.DefaultHeader(), Body: cmdtesting.ObjBody(codec, service)}, nil
			default:
				t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
				return nil, nil
			}
		}),
	}
	ioStreams, _, buf, _ := genericclioptions.NewTestIOStreams()
	cmd := NewCmdCreateServiceExternalName(tf, ioStreams)
	cmd.Flags().Set("output", "name")
	cmd.Flags().Set("external-name", "name")
	cmd.Run(cmd, []string{service.Name})
	expectedOutput := "service/" + service.Name + "\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output: %s, but got: %s", expectedOutput, buf.String())
	}
}
