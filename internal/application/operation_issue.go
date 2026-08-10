package application

// OperationIssue is a bounded, non-sensitive explanation that can cross the
// desktop boundary. Raw runtime and provisioning errors remain backend-only.
type OperationIssue struct {
	Code       string `json:"code"`
	Summary    string `json:"summary"`
	NextAction string `json:"nextAction"`
}

func installationIssue() *OperationIssue {
	return &OperationIssue{
		Code:       "application_install_failed",
		Summary:    "Não foi possível baixar ou iniciar o aplicativo.",
		NextAction: "Verifique a conexão e o espaço disponível, depois tente novamente.",
	}
}

func configurationIssue() *OperationIssue {
	return &OperationIssue{
		Code:       "application_configuration_failed",
		Summary:    "O aplicativo foi iniciado, mas a configuração automática não terminou.",
		NextAction: "Tente novamente. Se o problema continuar, abra o aplicativo para concluir a configuração.",
	}
}

func updateRollbackIssue() *OperationIssue {
	return &OperationIssue{
		Code:       "application_update_rolled_back",
		Summary:    "A nova versão não passou na verificação e a versão anterior foi restaurada.",
		NextAction: "O aplicativo pode continuar sendo usado. Tente a atualização novamente mais tarde.",
	}
}

func updateFailureIssue() *OperationIssue {
	return &OperationIssue{
		Code:       "application_update_failed",
		Summary:    "A atualização não terminou e o aplicativo precisa de atenção.",
		NextAction: "Não remova os dados. Exporte um diagnóstico antes de tentar novamente.",
	}
}

func statusUnavailableIssue() *OperationIssue {
	return &OperationIssue{
		Code:       "application_status_unavailable",
		Summary:    "Não foi possível verificar o estado deste aplicativo.",
		NextAction: "Verifique o ambiente e tente novamente. Os outros aplicativos não serão alterados.",
	}
}
