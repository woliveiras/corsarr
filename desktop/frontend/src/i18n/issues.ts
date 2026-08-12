const en = {
  'issue.application_configuration_failed.summary': 'The application could not be configured.',
  'issue.application_configuration_failed.next': 'Check that the service is running and try again.',
  'issue.application_install_failed.summary': 'The application could not be installed.',
  'issue.application_install_failed.next': 'Check the connection and try again.',
  'issue.application_status_unavailable.summary': 'The application status could not be checked.',
  'issue.application_status_unavailable.next': 'Check the environment and try again.',
  'issue.application_update_failed.summary': 'The application could not be updated.',
  'issue.application_update_failed.next':
    'The existing installation was preserved. Try again later.',
  'issue.application_update_rolled_back.summary': 'The new version failed verification.',
  'issue.application_update_rolled_back.next':
    'The previous version and configuration backup were preserved.',
  'issue.installation_failed.summary': 'Installation or configuration did not finish.',
  'issue.installation_failed.next': 'Try again. Completed steps will be reused.',
  'issue.onboarding_completion_failed.summary': 'Initial setup could not be completed.',
  'issue.onboarding_completion_failed.next': 'Your choices were preserved. Try again.',
  'issue.post_quality_configuration_failed.summary':
    'The applications are running, but final configuration did not finish.',
  'issue.post_quality_configuration_failed.next':
    'Try finalization again without reinstalling the applications.',
  'issue.quality_profile_sync_failed.summary': 'The quality profile could not be applied.',
  'issue.quality_profile_sync_failed.next': 'Check Radarr and Sonarr, then try again.',
  'issue.runtime_storage_access_denied.summary': 'Docker cannot access the selected folder.',
  'issue.runtime_storage_access_denied.next':
    'Choose an accessible folder or allow Docker to access it, then try again.',
} as const;

type IssueCatalog = Record<keyof typeof en, string>;

const es: IssueCatalog = {
  'issue.application_configuration_failed.summary': 'No se pudo configurar la aplicación.',
  'issue.application_configuration_failed.next':
    'Comprueba que el servicio esté activo e inténtalo de nuevo.',
  'issue.application_install_failed.summary': 'No se pudo instalar la aplicación.',
  'issue.application_install_failed.next': 'Comprueba la conexión e inténtalo de nuevo.',
  'issue.application_status_unavailable.summary':
    'No se pudo comprobar el estado de la aplicación.',
  'issue.application_status_unavailable.next': 'Comprueba el entorno e inténtalo de nuevo.',
  'issue.application_update_failed.summary': 'No se pudo actualizar la aplicación.',
  'issue.application_update_failed.next':
    'Se conservó la instalación existente. Inténtalo más tarde.',
  'issue.application_update_rolled_back.summary': 'La nueva versión no superó la comprobación.',
  'issue.application_update_rolled_back.next':
    'Se conservaron la versión anterior y la copia de la configuración.',
  'issue.installation_failed.summary': 'La instalación o configuración no terminó.',
  'issue.installation_failed.next': 'Inténtalo de nuevo. Se reutilizarán los pasos terminados.',
  'issue.onboarding_completion_failed.summary': 'No se pudo terminar la configuración inicial.',
  'issue.onboarding_completion_failed.next': 'Se conservaron tus elecciones. Inténtalo de nuevo.',
  'issue.post_quality_configuration_failed.summary':
    'Las aplicaciones están activas, pero la configuración final no terminó.',
  'issue.post_quality_configuration_failed.next':
    'Reintenta la finalización sin reinstalar las aplicaciones.',
  'issue.quality_profile_sync_failed.summary': 'No se pudo aplicar el perfil de calidad.',
  'issue.quality_profile_sync_failed.next': 'Comprueba Radarr y Sonarr e inténtalo de nuevo.',
  'issue.runtime_storage_access_denied.summary':
    'Docker no puede acceder a la carpeta seleccionada.',
  'issue.runtime_storage_access_denied.next':
    'Elige una carpeta accesible o permite que Docker acceda a ella e inténtalo de nuevo.',
};

const ptBR: IssueCatalog = {
  'issue.application_configuration_failed.summary': 'Não foi possível configurar o aplicativo.',
  'issue.application_configuration_failed.next':
    'Verifique se o serviço está rodando e tente novamente.',
  'issue.application_install_failed.summary': 'Não foi possível instalar o aplicativo.',
  'issue.application_install_failed.next': 'Verifique a conexão e tente novamente.',
  'issue.application_status_unavailable.summary':
    'Não foi possível verificar o estado do aplicativo.',
  'issue.application_status_unavailable.next': 'Verifique o ambiente e tente novamente.',
  'issue.application_update_failed.summary': 'Não foi possível atualizar o aplicativo.',
  'issue.application_update_failed.next':
    'A instalação existente foi preservada. Tente novamente mais tarde.',
  'issue.application_update_rolled_back.summary': 'A nova versão não passou na verificação.',
  'issue.application_update_rolled_back.next':
    'A versão anterior e o backup das configurações foram preservados.',
  'issue.installation_failed.summary': 'A instalação ou configuração não terminou.',
  'issue.installation_failed.next': 'Tente novamente. As etapas concluídas serão reaproveitadas.',
  'issue.onboarding_completion_failed.summary': 'Não foi possível concluir a configuração inicial.',
  'issue.onboarding_completion_failed.next': 'Suas escolhas foram preservadas. Tente novamente.',
  'issue.post_quality_configuration_failed.summary':
    'Os aplicativos estão rodando, mas a configuração final não terminou.',
  'issue.post_quality_configuration_failed.next':
    'Tente finalizar novamente sem reinstalar os aplicativos.',
  'issue.quality_profile_sync_failed.summary': 'Não foi possível aplicar o perfil de qualidade.',
  'issue.quality_profile_sync_failed.next': 'Verifique Radarr e Sonarr e tente novamente.',
  'issue.runtime_storage_access_denied.summary':
    'O Docker não consegue acessar a pasta selecionada.',
  'issue.runtime_storage_access_denied.next':
    'Escolha uma pasta acessível ou permita o acesso do Docker e tente novamente.',
};

const it: IssueCatalog = {
  'issue.application_configuration_failed.summary': 'Impossibile configurare l’applicazione.',
  'issue.application_configuration_failed.next': 'Verifica che il servizio sia attivo e riprova.',
  'issue.application_install_failed.summary': 'Impossibile installare l’applicazione.',
  'issue.application_install_failed.next': 'Verifica la connessione e riprova.',
  'issue.application_status_unavailable.summary':
    'Impossibile verificare lo stato dell’applicazione.',
  'issue.application_status_unavailable.next': 'Verifica l’ambiente e riprova.',
  'issue.application_update_failed.summary': 'Impossibile aggiornare l’applicazione.',
  'issue.application_update_failed.next':
    'L’installazione esistente è stata conservata. Riprova più tardi.',
  'issue.application_update_rolled_back.summary': 'La nuova versione non ha superato la verifica.',
  'issue.application_update_rolled_back.next':
    'La versione precedente e il backup della configurazione sono stati conservati.',
  'issue.installation_failed.summary': 'L’installazione o la configurazione non è terminata.',
  'issue.installation_failed.next': 'Riprova. I passaggi completati verranno riutilizzati.',
  'issue.onboarding_completion_failed.summary':
    'Impossibile completare la configurazione iniziale.',
  'issue.onboarding_completion_failed.next': 'Le tue scelte sono state conservate. Riprova.',
  'issue.post_quality_configuration_failed.summary':
    'Le applicazioni sono attive, ma la configurazione finale non è terminata.',
  'issue.post_quality_configuration_failed.next':
    'Riprova la finalizzazione senza reinstallare le applicazioni.',
  'issue.quality_profile_sync_failed.summary': 'Impossibile applicare il profilo di qualità.',
  'issue.quality_profile_sync_failed.next': 'Verifica Radarr e Sonarr, quindi riprova.',
  'issue.runtime_storage_access_denied.summary':
    'Docker non può accedere alla cartella selezionata.',
  'issue.runtime_storage_access_denied.next':
    'Scegli una cartella accessibile o consenti a Docker di accedervi, quindi riprova.',
};

export const issueMessages = { en, es, 'pt-BR': ptBR, it } as const;
