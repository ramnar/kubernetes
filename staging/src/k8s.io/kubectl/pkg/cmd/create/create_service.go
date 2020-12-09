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
	"context"
	"fmt"
	"strconv"
	"strings"
	
	"github.com/spf13/cobra"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
	utilsnet "k8s.io/utils/net"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// NewCmdCreateService is a macro command to create a new service
func NewCmdCreateService(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"svc"},
		Short:   i18n.T("Create a service using specified subcommand."),
		Long:    i18n.T("Create a service using specified subcommand."),
		Run:     cmdutil.DefaultSubCommandRun(ioStreams.ErrOut),
	}
	cmd.AddCommand(NewCmdCreateServiceClusterIP(f, ioStreams))
	cmd.AddCommand(NewCmdCreateServiceNodePort(f, ioStreams))
	cmd.AddCommand(NewCmdCreateServiceLoadBalancer(f, ioStreams))
	cmd.AddCommand(NewCmdCreateServiceExternalName(f, ioStreams))

	return cmd
}

var (
	serviceClusterIPLong = templates.LongDesc(i18n.T(`
    Create a ClusterIP service with the specified name.`))

	serviceClusterIPExample = templates.Examples(i18n.T(`
    # Create a new ClusterIP service named my-cs
    kubectl create service clusterip my-cs --tcp=5678:8080

    # Create a new ClusterIP service named my-cs (in headless mode)
    kubectl create service clusterip my-cs --clusterip="None"`))
)

func addPortFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("tcp", []string{}, "Port pairs can be specified as '<port>:<targetPort>'.")
}

// ServiceClusterIPOpts holds the options for 'create service clusterip' sub command
type ServiceCommon struct {
	// PrintFlags holds options necessary for obtaining a printer
    PrintFlags *genericclioptions.PrintFlags
	PrintObj   func(obj runtime.Object) error
	// Name of resource being created
	Name             string
	TCP          []string
	Type         v1.ServiceType
	ClusterIP    string
	NodePort     int
	ExternalName string
	
	DryRunStrategy   cmdutil.DryRunStrategy
	DryRunVerifier   *resource.DryRunVerifier
	CreateAnnotation bool
	FieldManager     string

	Namespace        string
	EnforceNamespace bool

	Mapper meta.RESTMapper
	Client *coreclient.CoreV1Client
	
	genericclioptions.IOStreams
	
}

// ServiceClusterIPOpts holds the options for 'create service clusterip' sub command
type ServiceClusterIPOpts struct {
     ServiceCommon   
}

// NewCmdCreateServiceClusterIP is a command to create a ClusterIP service
func NewCmdCreateServiceClusterIP(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := &ServiceClusterIPOpts{
		ServiceCommon{
			PrintFlags: genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
			IOStreams: ioStreams,
			Type: v1.ServiceTypeClusterIP,
		},
		}

	cmd := &cobra.Command{
		Use:                   "clusterip NAME [--tcp=<port>:<targetPort>] [--dry-run=server|client|none]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Create a ClusterIP service."),
		Long:                  serviceClusterIPLong,
		Example:               serviceClusterIPExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

    o.PrintFlags.AddFlags(cmd)
    
	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddDryRunFlag(cmd)
	addPortFlags(cmd)
	cmd.Flags().String("clusterip", "", i18n.T("Assign your own ClusterIP or set to 'None' for a 'headless' service (no loadbalancing)."))
	cmdutil.AddFieldManagerFlagVar(cmd, &o.FieldManager, "kubectl-create")
	return cmd
}

func errUnsupportedGenerator(cmd *cobra.Command, generatorName string) error {
	return cmdutil.UsageErrorf(cmd, "Generator %s not supported. ", generatorName)
}

// Complete completes all the required options
func (o *ServiceCommon) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	var err error
	o.Name, err = NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}

	restConfig, err := f.ToRESTConfig()
	if err != nil {
		return err
	}
	o.Client, err = coreclient.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	o.CreateAnnotation = cmdutil.GetFlagBool(cmd, cmdutil.ApplyAnnotationsFlag)

	o.DryRunStrategy, err = cmdutil.GetDryRunStrategy(cmd)
	if err != nil {
		return err
	}
	dynamicClient, err := f.DynamicClient()
	if err != nil {
		return err
	}
	discoveryClient, err := f.ToDiscoveryClient()
	if err != nil {
		return err
	}
	o.DryRunVerifier = resource.NewDryRunVerifier(dynamicClient, discoveryClient)

	o.Namespace, o.EnforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	cmdutil.PrintFlagsWithDryRunStrategy(o.PrintFlags, o.DryRunStrategy)

	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}

	o.PrintObj = func(obj runtime.Object) error {
		return printer.PrintObj(obj, o.Out)
	}

	return nil
}

//Complete completes all the required options
func (o *ServiceClusterIPOpts) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {

	o.TCP = cmdutil.GetFlagStringSlice(cmd, "tcp")
	o.ClusterIP = cmdutil.GetFlagString(cmd, "clusterip")
	
	return o.ServiceCommon.Complete(f, cmd, args)
	
}

// Validate checks to the ServiceClusterIPOpts to see if there is sufficient information run the command.
func (o *ServiceCommon) Validate() error {
	if len(o.Name) == 0 {
		return fmt.Errorf("name must be specified")
	}
	
	if len(o.Type) == 0 {
		return fmt.Errorf("type must be specified")
	}
	if o.ClusterIP == v1.ClusterIPNone && o.Type != v1.ServiceTypeClusterIP {
		return fmt.Errorf("ClusterIP=None can only be used with ClusterIP service type")
	}
	if o.ClusterIP != v1.ClusterIPNone && len(o.TCP) == 0 && o.Type != v1.ServiceTypeExternalName {
		return fmt.Errorf("at least one tcp port specifier must be provided")
	}
	if o.Type == v1.ServiceTypeExternalName {
		if errs := validation.IsDNS1123Subdomain(o.ExternalName); len(errs) != 0 {
			return fmt.Errorf("invalid service external name %s", o.ExternalName)
		}
	}
	return nil
}

// Run calls the CreateSubcommandOptions.Run in ServiceClusterIPOpts instance
func (o *ServiceCommon) Run() error {
	service, err := o.createService()
	if err != nil {
		return err
	}

	if err := util.CreateOrUpdateAnnotation(o.CreateAnnotation, service, scheme.DefaultJSONEncoder()); err != nil {
		return err
	}

	if o.DryRunStrategy != cmdutil.DryRunClient {
		createOptions := metav1.CreateOptions{}
		if o.FieldManager != "" {
			createOptions.FieldManager = o.FieldManager
		}
		if o.DryRunStrategy == cmdutil.DryRunServer {
			if err := o.DryRunVerifier.HasSupport(service.GroupVersionKind()); err != nil {
				return err
			}
			createOptions.DryRun = []string{metav1.DryRunAll}
		}
		service, err = o.Client.Services(o.Namespace).Create(context.TODO(), service, createOptions)
		if err != nil {
			return fmt.Errorf("failed to create service: %v", err)
		}
	}
	return o.PrintObj(service)
}

func parsePorts(portString string) (int32, intstr.IntOrString, error) {
	portStringSlice := strings.Split(portString, ":")
	port, err := utilsnet.ParsePort(portStringSlice[0], true)
	if err != nil {
		return 0, intstr.FromInt(0), err
	}

	if len(portStringSlice) == 1 {
		return int32(port), intstr.FromInt(int(port)), nil
	}
	var targetPort intstr.IntOrString
	if portNum, err := strconv.Atoi(portStringSlice[1]); err != nil {
		if errs := validation.IsValidPortName(portStringSlice[1]); len(errs) != 0 {
			return 0, intstr.FromInt(0), fmt.Errorf(strings.Join(errs, ","))
		}
		targetPort = intstr.FromString(portStringSlice[1])
	} else {
		if errs := validation.IsValidPortNum(portNum); len(errs) != 0 {
			return 0, intstr.FromInt(0), fmt.Errorf(strings.Join(errs, ","))
		}
		targetPort = intstr.FromInt(portNum)
	}
	return int32(port), targetPort, nil
}

func (o *ServiceCommon) createService() (*v1.Service, error) {
	namespace := ""
	if o.EnforceNamespace {
		namespace = o.Namespace
	}

	ports := []v1.ServicePort{}
	for _, tcpString := range o.TCP {
		port, targetPort, err := parsePorts(tcpString)
		if err != nil {
			return nil, err
		}

		portName := strings.Replace(tcpString, ":", "-", -1)
		ports = append(ports, v1.ServicePort{
			Name:       portName,
			Port:       port,
			TargetPort: targetPort,
			Protocol:   v1.Protocol("TCP"),
			NodePort:   int32(o.NodePort),
		})
	}

	// setup default label and selector
	labels := map[string]string{}
	labels["app"] = o.Name
	selector := map[string]string{}
	selector["app"] = o.Name

	service := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   o.Name,
			Namespace: namespace,
			Labels: labels,
		},
		Spec: v1.ServiceSpec{
			Type:         v1.ServiceType(o.Type),
			Selector:     selector,
			Ports:        ports,
			ExternalName: o.ExternalName,
		},
	}
	if len(o.ClusterIP) > 0 {
		service.Spec.ClusterIP = o.ClusterIP
	}

	return &service, nil
}

var (
	serviceNodePortLong = templates.LongDesc(i18n.T(`
    Create a NodePort service with the specified name.`))

	serviceNodePortExample = templates.Examples(i18n.T(`
    # Create a new NodePort service named my-ns
    kubectl create service nodeport my-ns --tcp=5678:8080`))
)

// ServiceNodePortOpts holds the options for 'create service nodeport' sub command
type ServiceNodePortOpts struct {
	ServiceCommon
}

// NewCmdCreateServiceNodePort is a macro command for creating a NodePort service
func NewCmdCreateServiceNodePort(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := &ServiceNodePortOpts{
		ServiceCommon{
			PrintFlags: genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
			IOStreams: ioStreams,
			Type: v1.ServiceTypeNodePort,
			ClusterIP: "",
		},
	}

	cmd := &cobra.Command{
		Use:                   "nodeport NAME [--tcp=port:targetPort] [--dry-run=server|client|none]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Create a NodePort service."),
		Long:                  serviceNodePortLong,
		Example:               serviceNodePortExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	o.PrintFlags.AddFlags(cmd)

	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddDryRunFlag(cmd)
	cmd.Flags().Int("node-port", 0, "Port used to expose the service on each node in a cluster.")
	cmdutil.AddFieldManagerFlagVar(cmd, &o.FieldManager, "kubectl-create")
	addPortFlags(cmd)
	return cmd
}

// Complete completes all the required options
func (o *ServiceNodePortOpts) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {			
	o.TCP = cmdutil.GetFlagStringSlice(cmd, "tcp")
	o.NodePort = cmdutil.GetFlagInt(cmd, "node-port")
	return o.ServiceCommon.Complete(f, cmd, args)
}

// Run calls the CreateSubcommandOptions.Run in ServiceNodePortOpts instance
func (o *ServiceNodePortOpts) Run() error {
	return o.ServiceCommon.Run()
}

var (
	serviceLoadBalancerLong = templates.LongDesc(i18n.T(`
    Create a LoadBalancer service with the specified name.`))

	serviceLoadBalancerExample = templates.Examples(i18n.T(`
    # Create a new LoadBalancer service named my-lbs
    kubectl create service loadbalancer my-lbs --tcp=5678:8080`))
)

// ServiceLoadBalancerOpts holds the options for 'create service loadbalancer' sub command
type ServiceLoadBalancerOpts struct {
	ServiceCommon
}

// NewCmdCreateServiceLoadBalancer is a macro command for creating a LoadBalancer service
func NewCmdCreateServiceLoadBalancer(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := &ServiceLoadBalancerOpts{
		ServiceCommon{
			PrintFlags: genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
			IOStreams: ioStreams,
			Type: v1.ServiceTypeLoadBalancer,
			ClusterIP: "",
		},
	}

	cmd := &cobra.Command{
		Use:                   "loadbalancer NAME [--tcp=port:targetPort] [--dry-run=server|client|none]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Create a LoadBalancer service."),
		Long:                  serviceLoadBalancerLong,
		Example:               serviceLoadBalancerExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	o.PrintFlags.AddFlags(cmd)

	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddDryRunFlag(cmd)
	
	cmdutil.AddFieldManagerFlagVar(cmd, &o.FieldManager, "kubectl-create")
	addPortFlags(cmd)
	return cmd
}

// Complete completes all the required options
func (o *ServiceLoadBalancerOpts) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	o.TCP = cmdutil.GetFlagStringSlice(cmd, "tcp")
	return o.ServiceCommon.Complete(f, cmd, args)
}

// Run calls the CreateSubcommandOptions.Run in ServiceLoadBalancerOpts instance
func (o *ServiceLoadBalancerOpts) Run() error {
	return o.ServiceCommon.Run()
}

var (
	serviceExternalNameLong = templates.LongDesc(i18n.T(`
	Create an ExternalName service with the specified name.

	ExternalName service references to an external DNS address instead of
	only pods, which will allow application authors to reference services
	that exist off platform, on other clusters, or locally.`))

	serviceExternalNameExample = templates.Examples(i18n.T(`
	# Create a new ExternalName service named my-ns
	kubectl create service externalname my-ns --external-name bar.com`))
)

// ServiceExternalNameOpts holds the options for 'create service externalname' sub command
type ServiceExternalNameOpts struct {
	ServiceCommon
}

// NewCmdCreateServiceExternalName is a macro command for creating an ExternalName service
func NewCmdCreateServiceExternalName(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := &ServiceExternalNameOpts{
		ServiceCommon{
			PrintFlags: genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
			IOStreams: ioStreams,
			Type: v1.ServiceTypeExternalName,
			ClusterIP: "",
		},
	}

	cmd := &cobra.Command{
		Use:                   "externalname NAME --external-name external.name [--dry-run=server|client|none]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Create an ExternalName service."),
		Long:                  serviceExternalNameLong,
		Example:               serviceExternalNameExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	o.PrintFlags.AddFlags(cmd)

	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddDryRunFlag(cmd)
	addPortFlags(cmd)
	cmd.Flags().String("external-name", "", i18n.T("External name of service"))
	cmd.MarkFlagRequired("external-name")
	cmdutil.AddFieldManagerFlagVar(cmd, &o.FieldManager, "kubectl-create")
	return cmd
}

// Complete completes all the required options
func (o *ServiceExternalNameOpts) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	o.ExternalName = cmdutil.GetFlagString(cmd, "external-name")
	return o.ServiceCommon.Complete(f, cmd, args)
}

// Run calls the CreateSubcommandOptions.Run in ServiceExternalNameOpts instance
func (o *ServiceExternalNameOpts) Run() error {
	return o.ServiceCommon.Run()
}
