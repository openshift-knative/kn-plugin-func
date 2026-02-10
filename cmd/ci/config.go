package ci

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ory/viper"
	"knative.dev/func/cmd/common"
)

const (
	ConfigCIFeatureFlag = "FUNC_ENABLE_CI_CONFIG"

	PathFlag = "path"

	PlatformFlag    = "platform"
	DefaultPlatform = "github"

	DefaultGitHubWorkflowDir      = ".github/workflows"
	DefaultGitHubWorkflowFilename = "func-deploy.yaml"

	BranchFlag    = "branch"
	DefaultBranch = "main"

	WorkflowNameFlag               = "workflow-name"
	DefaultWorkflowName            = "Func Deploy"
	DefaultRemoteBuildWorkflowName = "Remote " + DefaultWorkflowName

	KubeconfigSecretNameFlag    = "kubeconfig-secret-name"
	DefaultKubeconfigSecretName = "KUBECONFIG"

	RegistryLoginUrlVariableNameFlag    = "registry-login-url-variable-name"
	DefaultRegistryLoginUrlVariableName = "REGISTRY_LOGIN_URL"

	RegistryUserVariableNameFlag    = "registry-user-variable-name"
	DefaultRegistryUserVariableName = "REGISTRY_USERNAME"

	RegistryPassSecretNameFlag    = "registry-pass-secret-name"
	DefaultRegistryPassSecretName = "REGISTRY_PASSWORD"

	RegistryUrlVariableNameFlag    = "registry-url-variable-name"
	DefaultRegistryUrlVariableName = "REGISTRY_URL"

	UseRegistryLoginFlag    = "use-registry-login"
	DefaultUseRegistryLogin = true

	WorkflowDispatchFlag    = "workflow-dispatch"
	DefaultWorkflowDispatch = false

	UseRemoteBuildFlag    = "remote"
	DefaultUseRemoteBuild = false

	UseSelfHostedRunnerFlag    = "self-hosted-runner"
	DefaultUseSelfHostedRunner = false

	ForceFlag    = "force"
	DefaultForce = false
)

// CIConfig readonly configuration
type CIConfig struct {
	githubWorkflowDir,
	githubWorkflowFilename,
	path,
	branch,
	workflowName,
	kubeconfigSecret,
	registryLoginUrlVar,
	registryUserVar,
	registryPassSecret,
	registryUrlVar string
	useRegistryLogin,
	useSelfHostedRunner,
	useRemoteBuild,
	useWorkflowDispatch,
	force bool
}

func NewCIConfig(
	currentBranch common.CurrentBranchFunc,
	workingDir common.WorkDirFunc,
	workflowNameExplicit bool,
) (CIConfig, error) {
	if err := resolvePlatform(); err != nil {
		return CIConfig{}, err
	}

	path, err := resolvePath(workingDir)
	if err != nil {
		return CIConfig{}, err
	}

	branch, err := resolveBranch(path, currentBranch)
	if err != nil {
		return CIConfig{}, err
	}

	workflowName := resolveWorkflowName(workflowNameExplicit)

	return CIConfig{
		githubWorkflowDir:      DefaultGitHubWorkflowDir,
		githubWorkflowFilename: DefaultGitHubWorkflowFilename,
		path:                   path,
		branch:                 branch,
		workflowName:           workflowName,
		kubeconfigSecret:       viper.GetString(KubeconfigSecretNameFlag),
		registryLoginUrlVar:    viper.GetString(RegistryLoginUrlVariableNameFlag),
		registryUserVar:        viper.GetString(RegistryUserVariableNameFlag),
		registryPassSecret:     viper.GetString(RegistryPassSecretNameFlag),
		registryUrlVar:         viper.GetString(RegistryUrlVariableNameFlag),
		useRegistryLogin:       viper.GetBool(UseRegistryLoginFlag),
		useSelfHostedRunner:    viper.GetBool(UseSelfHostedRunnerFlag),
		useRemoteBuild:         viper.GetBool(UseRemoteBuildFlag),
		useWorkflowDispatch:    viper.GetBool(WorkflowDispatchFlag),
		force:                  viper.GetBool(ForceFlag),
	}, nil
}

func resolvePlatform() error {
	platform := viper.GetString(PlatformFlag)
	if strings.ToLower(platform) != DefaultPlatform {
		return fmt.Errorf("%s support is not implemented", platform)
	}

	return nil
}

func resolvePath(workingDir common.WorkDirFunc) (string, error) {
	path := viper.GetString(PathFlag)
	if path != "" && path != "." {
		return path, nil
	}

	cwd, err := workingDir()
	if err != nil {
		return "", err
	}

	return cwd, nil
}

func resolveBranch(path string, currentBranch common.CurrentBranchFunc) (string, error) {
	branch := viper.GetString(BranchFlag)
	if branch != "" {
		return branch, nil
	}

	branch, err := currentBranch(path)
	if err != nil {
		return "", err
	}

	return branch, nil
}

func resolveWorkflowName(explicit bool) string {
	workflowName := viper.GetString(WorkflowNameFlag)
	if explicit {
		return workflowName
	}

	if viper.GetBool(UseRemoteBuildFlag) {
		return DefaultRemoteBuildWorkflowName
	}

	return DefaultWorkflowName
}

func (cc *CIConfig) FnGitHubWorkflowDir(fnRoot string) string {
	return filepath.Join(fnRoot, cc.githubWorkflowDir)
}

func (cc *CIConfig) FnGitHubWorkflowFilepath(fnRoot string) string {
	return filepath.Join(cc.FnGitHubWorkflowDir(fnRoot), cc.githubWorkflowFilename)
}

func (cc *CIConfig) Path() string {
	return cc.path
}

func (cc *CIConfig) WorkflowName() string {
	return cc.workflowName
}

func (cc *CIConfig) Branch() string {
	return cc.branch
}

func (cc *CIConfig) UseRegistryLogin() bool {
	return cc.useRegistryLogin
}

func (cc *CIConfig) UseSelfHostedRunner() bool {
	return cc.useSelfHostedRunner
}

func (cc *CIConfig) UseRemoteBuild() bool {
	return cc.useRemoteBuild
}

func (cc *CIConfig) UseWorkflowDispatch() bool {
	return cc.useWorkflowDispatch
}

func (cc *CIConfig) KubeconfigSecret() string {
	return cc.kubeconfigSecret
}

func (cc *CIConfig) RegistryLoginUrlVar() string {
	return cc.registryLoginUrlVar
}

func (cc *CIConfig) RegistryUserVar() string {
	return cc.registryUserVar
}

func (cc *CIConfig) RegistryPassSecret() string {
	return cc.registryPassSecret
}

func (cc *CIConfig) RegistryUrlVar() string {
	return cc.registryUrlVar
}

func (cc *CIConfig) Force() bool {
	return cc.force
}
