export const operationMessages = {
  en: {
    'site.openError': 'Could not open the official website.',
    'app.installAria': 'Install {{name}}',
    'app.installedTitle': 'This application is already installed.',
    'app.installedAria': '{{name}} is installed.',
    'app.environmentBlocked': 'Check the environment before trying to install again.',
    'app.environmentBlockedAria': '{{name}} cannot be installed until the environment is verified.',
    'app.manualTitle':
      'This application does not yet have secure automatic configuration in Corsarr.',
    'app.manualAria': '{{name}} does not yet have automatic installation in Corsarr.',
    'app.preparingInstall': 'Preparing to install {{name}}…',
    'app.installIncomplete': 'The installation of {{name}} did not finish.',
    'app.tryAgain': 'Try again.',
    'app.installReady': '{{name}} was installed and is ready to use.',
    'app.installError': 'Could not install {{name}}. Try again.',
    'app.openAria': 'Open {{name}} in the browser',
    'app.opening': 'Opening {{name}}…',
    'app.opened': '{{name}} was opened in the browser.',
    'app.openError': 'Could not open {{name}}. The service may not be installed yet.',
    'app.updateConfirm':
      'Update {{name}}? Corsarr will create a private configuration backup and verify the new version before completing. The application will be unavailable briefly. If the version migrates the database, restoring the previous image may not undo that migration.',
    'app.updating': 'Updating…',
    'app.updatePreparing': 'Creating a backup and checking the {{name}} update…',
    'app.updateReady': '{{name}} was updated and verified. The configuration backup was preserved.',
    'app.updateRolledBack':
      'The new version of {{name}} failed verification. The previous image was restored and the backup was preserved.',
    'app.updateAttention':
      '{{name}} needs attention after the update attempt. See technical details.',
    'app.updateCurrent': '{{name}} already uses the version approved by Corsarr.',
    'app.updateError':
      'Could not start the {{name}} update. No changes were authorized outside Corsarr resources.',
    'app.lifecycleError': 'Could not {{action}} the application.',
    'app.removeConfirm': 'Remove {{name}}? Its configuration and your media will be preserved.',
    'app.removeFirst': 'Remove first: {{names}}.',
    'app.removeBlockedAria': 'Cannot remove {{name}}. {{reason}}',
    'app.removeBlocked': 'To remove {{name}}, first remove: {{names}}.',
    'app.removeDataConfirm':
      'Remove approximately {{size}} of {{name}} configuration? The library and downloads will not change. The configuration will be moved to the Corsarr trash and can be recovered manually.',
    'app.dataArchived': '{{name}} configuration was moved to the Corsarr trash.',
    'app.noData': '{{name}} has no configuration to remove.',
    'app.removeDataError':
      'Could not remove {{name}} configuration. Confirm that the application has already been removed.',
    'app.copyPasswordError': 'Could not copy the {{name}} password.',
    'app.addressCopied': 'Address copied. Open it on a device connected to the same local network.',
    'app.addressCopyError':
      'Could not copy the address. Check that local-network access remains enabled.',
    'storage.saved': 'Folder saved',
    'storage.savedDescription':
      'Corsarr will check this folder again before preparing applications.',
    'storage.savedBadge': 'Saved',
    'storage.changeFolder': 'Change folder',
    'setup.loadError': 'Could not restore the preparation saved on this computer.',
    'environment.readyTitle': 'Environment ready',
    'environment.readyDescription': 'This computer is ready for Corsarr applications.',
    'environment.readyBadge': 'Ready',
    'environment.requiredTitle': 'Preparation required',
    'environment.requiredDescription':
      'Corsarr can prepare the required components after your explicit authorization.',
    'environment.requiredBadge': 'Not prepared',
    'environment.stoppedTitle': 'Environment stopped',
    'environment.stoppedDescription': 'The required component is installed but must be started.',
    'environment.attentionTitle': 'Could not verify',
    'environment.attentionDescription':
      'Try again. Technical details may help diagnose the problem.',
    'environment.attentionBadge': 'Attention',
    'environment.start': 'Start environment',
    'environment.version': 'Version: {{version}}',
    'environment.diagnostic': 'Diagnostics: {{detail}}',
    'environment.memory': 'Memory: {{amount}} GiB',
    'environment.disk': 'Free space: {{amount}} GiB',
    'environment.stoppedBadge': 'Stopped',
    'environment.hostAttention': 'This computer needs attention',
    'environment.requirements': 'Requirements',
    'environment.provider': 'Provider: {{provider}}',
    'environment.state': 'State: {{state}}',
    'environment.retry': 'Try again in a few moments.',
    'runtime.resolveRequirements':
      'Resolve this computer’s listed requirements before preparing the environment.',
    'runtime.authorizationUnavailable':
      'Initial authorization is unavailable. Restart setup to prepare this computer.',
    'runtime.installConfirm':
      'Corsarr will download Docker Desktop 4.86.0 directly from Docker, verify its checksum and official signature, and request macOS authorization to install it. Continue?',
    'runtime.installing': 'Installing…',
    'runtime.starting': 'Starting…',
    'runtime.downloading':
      'Downloading and verifying official components. macOS may request your password.',
    'runtime.startingDescription': 'Starting the required components. This may take a few moments.',
    'runtime.installedReady': 'This computer was prepared and is ready to install applications.',
    'runtime.startedReady': 'The environment was started and is ready.',
    'runtime.incomplete':
      'Preparation did not finish. Nothing was installed without a valid signature; try again or see technical details.',
    'runtime.error': 'Could not finish preparing this computer.',
  },
  es: {
    'site.openError': 'No se pudo abrir el sitio oficial.',
    'app.installAria': 'Instalar {{name}}',
    'app.installedTitle': 'Esta aplicación ya está instalada.',
    'app.installedAria': '{{name}} está instalada.',
    'app.environmentBlocked': 'Comprueba el entorno antes de intentar instalar de nuevo.',
    'app.environmentBlockedAria': '{{name}} no se puede instalar hasta comprobar el entorno.',
    'app.manualTitle': 'Esta aplicación aún no tiene configuración automática segura en Corsarr.',
    'app.manualAria': '{{name}} aún no tiene instalación automática en Corsarr.',
    'app.preparingInstall': 'Preparando la instalación de {{name}}…',
    'app.installIncomplete': 'La instalación de {{name}} no terminó.',
    'app.tryAgain': 'Inténtalo de nuevo.',
    'app.installReady': '{{name}} se instaló y está lista para usar.',
    'app.installError': 'No se pudo instalar {{name}}. Inténtalo de nuevo.',
    'app.openAria': 'Abrir {{name}} en el navegador',
    'app.opening': 'Abriendo {{name}}…',
    'app.opened': '{{name}} se abrió en el navegador.',
    'app.openError': 'No se pudo abrir {{name}}. Es posible que el servicio aún no esté instalado.',
    'app.updateConfirm':
      '¿Actualizar {{name}}? Corsarr creará una copia privada de la configuración y comprobará la nueva versión antes de terminar. La aplicación no estará disponible durante unos instantes. Si la versión migra la base de datos, restaurar la imagen anterior puede no deshacer esa migración.',
    'app.updating': 'Actualizando…',
    'app.updatePreparing': 'Creando una copia y comprobando la actualización de {{name}}…',
    'app.updateReady':
      '{{name}} se actualizó y comprobó. Se conservó la copia de la configuración.',
    'app.updateRolledBack':
      'La nueva versión de {{name}} no superó la comprobación. Se restauró la imagen anterior y se conservó la copia.',
    'app.updateAttention':
      '{{name}} necesita atención tras el intento de actualización. Consulta los detalles técnicos.',
    'app.updateCurrent': '{{name}} ya usa la versión aprobada por Corsarr.',
    'app.updateError':
      'No se pudo iniciar la actualización de {{name}}. No se autorizaron cambios fuera de los recursos de Corsarr.',
    'app.lifecycleError': 'No se pudo {{action}} la aplicación.',
    'app.removeConfirm':
      '¿Eliminar {{name}}? Se conservarán su configuración y tus archivos multimedia.',
    'app.removeFirst': 'Elimina primero: {{names}}.',
    'app.removeBlockedAria': 'No se puede eliminar {{name}}. {{reason}}',
    'app.removeBlocked': 'Para eliminar {{name}}, elimina primero: {{names}}.',
    'app.removeDataConfirm':
      '¿Eliminar aproximadamente {{size}} de configuración de {{name}}? La biblioteca y las descargas no cambiarán. La configuración se moverá a la papelera de Corsarr y podrá recuperarse manualmente.',
    'app.dataArchived': 'La configuración de {{name}} se movió a la papelera de Corsarr.',
    'app.noData': '{{name}} no tiene configuración que eliminar.',
    'app.removeDataError':
      'No se pudo eliminar la configuración de {{name}}. Confirma que la aplicación ya se eliminó.',
    'app.copyPasswordError': 'No se pudo copiar la contraseña de {{name}}.',
    'app.addressCopied':
      'Dirección copiada. Ábrela en un dispositivo conectado a la misma red local.',
    'app.addressCopyError':
      'No se pudo copiar la dirección. Comprueba que el acceso por red local siga activado.',
    'storage.saved': 'Carpeta guardada',
    'storage.savedDescription':
      'Corsarr volverá a comprobar esta carpeta antes de preparar las aplicaciones.',
    'storage.savedBadge': 'Guardada',
    'storage.changeFolder': 'Cambiar carpeta',
    'setup.loadError': 'No se pudo recuperar la preparación guardada en este equipo.',
    'environment.readyTitle': 'Entorno listo',
    'environment.readyDescription': 'Este equipo está listo para las aplicaciones de Corsarr.',
    'environment.readyBadge': 'Listo',
    'environment.requiredTitle': 'Preparación necesaria',
    'environment.requiredDescription':
      'Corsarr puede preparar los componentes necesarios tras tu autorización expresa.',
    'environment.requiredBadge': 'Sin preparar',
    'environment.stoppedTitle': 'Entorno detenido',
    'environment.stoppedDescription':
      'El componente necesario está instalado, pero debe iniciarse.',
    'environment.attentionTitle': 'No se pudo comprobar',
    'environment.attentionDescription':
      'Inténtalo de nuevo. Los detalles técnicos pueden ayudar con el diagnóstico.',
    'environment.attentionBadge': 'Atención',
    'environment.start': 'Iniciar entorno',
    'environment.version': 'Versión: {{version}}',
    'environment.diagnostic': 'Diagnóstico: {{detail}}',
    'environment.memory': 'Memoria: {{amount}} GiB',
    'environment.disk': 'Espacio libre: {{amount}} GiB',
    'environment.stoppedBadge': 'Detenido',
    'environment.hostAttention': 'Este equipo necesita atención',
    'environment.requirements': 'Requisitos',
    'environment.provider': 'Proveedor: {{provider}}',
    'environment.state': 'Estado: {{state}}',
    'environment.retry': 'Inténtalo de nuevo en unos instantes.',
    'runtime.resolveRequirements':
      'Resuelve los requisitos indicados en este equipo antes de preparar el entorno.',
    'runtime.authorizationUnavailable':
      'La autorización inicial no está disponible. Reinicia la configuración para preparar este equipo.',
    'runtime.installConfirm':
      'Corsarr descargará Docker Desktop 4.86.0 directamente de Docker, comprobará el checksum y la firma oficial y solicitará autorización de macOS para instalarlo. ¿Continuar?',
    'runtime.installing': 'Instalando…',
    'runtime.starting': 'Iniciando…',
    'runtime.downloading':
      'Descargando y comprobando los componentes oficiales. macOS puede solicitar tu contraseña.',
    'runtime.startingDescription':
      'Iniciando los componentes necesarios. Esto puede tardar unos instantes.',
    'runtime.installedReady': 'Este equipo está preparado y listo para instalar aplicaciones.',
    'runtime.startedReady': 'El entorno se inició y está listo.',
    'runtime.incomplete':
      'La preparación no terminó. No se instaló nada sin una firma válida; inténtalo de nuevo o consulta los detalles técnicos.',
    'runtime.error': 'No se pudo terminar de preparar este equipo.',
  },
  'pt-BR': {
    'site.openError': 'Não foi possível abrir o site oficial.',
    'app.installAria': 'Instalar {{name}}',
    'app.installedTitle': 'Este aplicativo já está instalado.',
    'app.installedAria': '{{name}} está instalado.',
    'app.environmentBlocked': 'Verifique o ambiente antes de tentar instalar novamente.',
    'app.environmentBlockedAria':
      '{{name}} não pode ser instalado enquanto o ambiente não for verificado.',
    'app.manualTitle':
      'Este aplicativo ainda não possui configuração automática segura no Corsarr.',
    'app.manualAria': '{{name}} ainda não possui instalação automática no Corsarr.',
    'app.preparingInstall': 'Preparando a instalação de {{name}}…',
    'app.installIncomplete': 'A instalação de {{name}} não terminou.',
    'app.tryAgain': 'Tente novamente.',
    'app.installReady': '{{name}} foi instalado e está pronto para uso.',
    'app.installError': 'Não foi possível instalar {{name}}. Tente novamente.',
    'app.openAria': 'Abrir {{name}} no navegador',
    'app.opening': 'Abrindo {{name}}…',
    'app.opened': '{{name}} foi aberto no navegador.',
    'app.openError': 'Não foi possível abrir {{name}}. O serviço ainda pode não estar instalado.',
    'app.updateConfirm':
      'Atualizar {{name}}? O Corsarr criará um backup privado das configurações e verificará a nova versão antes de concluir. O aplicativo ficará indisponível por alguns instantes. Se a versão migrar o banco de dados, restaurar a imagem anterior pode não desfazer essa migração.',
    'app.updating': 'Atualizando…',
    'app.updatePreparing': 'Criando backup e verificando a atualização de {{name}}…',
    'app.updateReady':
      '{{name}} foi atualizado e verificado. O backup das configurações foi preservado.',
    'app.updateRolledBack':
      'A nova versão de {{name}} não passou na verificação. A imagem anterior foi restaurada e o backup foi preservado.',
    'app.updateAttention':
      '{{name}} requer atenção após a tentativa de atualização. Consulte os detalhes técnicos.',
    'app.updateCurrent': '{{name}} já usa a versão aprovada pelo Corsarr.',
    'app.updateError':
      'Não foi possível iniciar a atualização de {{name}}. Nenhuma alteração foi autorizada fora dos recursos do Corsarr.',
    'app.lifecycleError': 'Não foi possível {{action}} o aplicativo.',
    'app.removeConfirm': 'Remover {{name}}? As configurações e sua mídia serão preservadas.',
    'app.removeFirst': 'Remova primeiro: {{names}}.',
    'app.removeBlockedAria': 'Não é possível remover {{name}}. {{reason}}',
    'app.removeBlocked': 'Para remover {{name}}, remova primeiro: {{names}}.',
    'app.removeDataConfirm':
      'Remover aproximadamente {{size}} de configurações de {{name}}? A biblioteca e os downloads não serão alterados. A configuração será movida para a lixeira do Corsarr e poderá ser recuperada manualmente.',
    'app.dataArchived': 'As configurações de {{name}} foram movidas para a lixeira do Corsarr.',
    'app.noData': '{{name}} não possui configurações para remover.',
    'app.removeDataError':
      'Não foi possível remover as configurações de {{name}}. Confirme que o aplicativo já foi removido.',
    'app.copyPasswordError': 'Não foi possível copiar a senha do {{name}}.',
    'app.addressCopied': 'Endereço copiado. Abra-o em um aparelho conectado à mesma rede local.',
    'app.addressCopyError':
      'Não foi possível copiar o endereço. Verifique se o acesso pela rede continua ativado.',
    'storage.saved': 'Pasta salva',
    'storage.savedDescription':
      'O Corsarr verificará novamente esta pasta antes de preparar os aplicativos.',
    'storage.savedBadge': 'Salva',
    'storage.changeFolder': 'Trocar pasta',
    'setup.loadError': 'Não foi possível recuperar a preparação salva neste computador.',
    'environment.readyTitle': 'Ambiente pronto',
    'environment.readyDescription':
      'Este computador está pronto para receber os aplicativos do Corsarr.',
    'environment.readyBadge': 'Pronto',
    'environment.requiredTitle': 'Preparação necessária',
    'environment.requiredDescription':
      'O Corsarr pode preparar os componentes necessários após sua autorização explícita.',
    'environment.requiredBadge': 'Não preparado',
    'environment.stoppedTitle': 'Ambiente parado',
    'environment.stoppedDescription':
      'O componente necessário está instalado, mas precisa ser iniciado.',
    'environment.attentionTitle': 'Não foi possível verificar',
    'environment.attentionDescription':
      'Tente novamente. Os detalhes técnicos podem ajudar no diagnóstico.',
    'environment.attentionBadge': 'Atenção',
    'environment.start': 'Iniciar ambiente',
    'environment.version': 'Versão: {{version}}',
    'environment.diagnostic': 'Diagnóstico: {{detail}}',
    'environment.memory': 'Memória: {{amount}} GiB',
    'environment.disk': 'Espaço livre: {{amount}} GiB',
    'environment.stoppedBadge': 'Parado',
    'environment.hostAttention': 'Este computador precisa de atenção',
    'environment.requirements': 'Requisitos',
    'environment.provider': 'Provedor: {{provider}}',
    'environment.state': 'Estado: {{state}}',
    'environment.retry': 'Tente novamente em alguns instantes.',
    'runtime.resolveRequirements':
      'Resolva os requisitos indicados neste computador antes de preparar o ambiente.',
    'runtime.authorizationUnavailable':
      'A autorização inicial não está disponível. Reinicie a configuração para preparar este computador.',
    'runtime.installConfirm':
      'O Corsarr baixará o Docker Desktop 4.86.0 diretamente da Docker, verificará o checksum e a assinatura oficial e solicitará a autorização do macOS para instalar. Continuar?',
    'runtime.installing': 'Instalando…',
    'runtime.starting': 'Iniciando…',
    'runtime.downloading':
      'Baixando e verificando os componentes oficiais. O macOS poderá solicitar sua senha.',
    'runtime.startingDescription':
      'Iniciando os componentes necessários. Isso pode levar alguns instantes.',
    'runtime.installedReady':
      'Este computador foi preparado e está pronto para instalar os aplicativos.',
    'runtime.startedReady': 'O ambiente foi iniciado e está pronto.',
    'runtime.incomplete':
      'A preparação não terminou. Nada foi instalado sem assinatura válida; tente novamente ou consulte os detalhes técnicos.',
    'runtime.error': 'Não foi possível concluir a preparação deste computador.',
  },
  it: {
    'site.openError': 'Impossibile aprire il sito ufficiale.',
    'app.installAria': 'Installa {{name}}',
    'app.installedTitle': 'Questa applicazione è già installata.',
    'app.installedAria': '{{name}} è installata.',
    'app.environmentBlocked': 'Verifica l’ambiente prima di provare di nuovo l’installazione.',
    'app.environmentBlockedAria':
      '{{name}} non può essere installata finché l’ambiente non è verificato.',
    'app.manualTitle':
      'Questa applicazione non dispone ancora di una configurazione automatica sicura in Corsarr.',
    'app.manualAria': '{{name}} non dispone ancora dell’installazione automatica in Corsarr.',
    'app.preparingInstall': 'Preparazione dell’installazione di {{name}}…',
    'app.installIncomplete': 'L’installazione di {{name}} non è terminata.',
    'app.tryAgain': 'Riprova.',
    'app.installReady': '{{name}} è stata installata ed è pronta all’uso.',
    'app.installError': 'Impossibile installare {{name}}. Riprova.',
    'app.openAria': 'Apri {{name}} nel browser',
    'app.opening': 'Apertura di {{name}}…',
    'app.opened': '{{name}} è stata aperta nel browser.',
    'app.openError':
      'Impossibile aprire {{name}}. Il servizio potrebbe non essere ancora installato.',
    'app.updateConfirm':
      'Aggiornare {{name}}? Corsarr creerà un backup privato della configurazione e verificherà la nuova versione prima di terminare. L’applicazione non sarà disponibile per alcuni istanti. Se la versione migra il database, ripristinare l’immagine precedente potrebbe non annullare la migrazione.',
    'app.updating': 'Aggiornamento…',
    'app.updatePreparing': 'Creazione del backup e verifica dell’aggiornamento di {{name}}…',
    'app.updateReady':
      '{{name}} è stata aggiornata e verificata. Il backup della configurazione è stato conservato.',
    'app.updateRolledBack':
      'La nuova versione di {{name}} non ha superato la verifica. È stata ripristinata l’immagine precedente e il backup è stato conservato.',
    'app.updateAttention':
      '{{name}} richiede attenzione dopo il tentativo di aggiornamento. Consulta i dettagli tecnici.',
    'app.updateCurrent': '{{name}} usa già la versione approvata da Corsarr.',
    'app.updateError':
      'Impossibile avviare l’aggiornamento di {{name}}. Non sono state autorizzate modifiche fuori dalle risorse di Corsarr.',
    'app.lifecycleError': 'Impossibile {{action}} l’applicazione.',
    'app.removeConfirm':
      'Rimuovere {{name}}? La configurazione e i contenuti multimediali saranno conservati.',
    'app.removeFirst': 'Rimuovi prima: {{names}}.',
    'app.removeBlockedAria': 'Impossibile rimuovere {{name}}. {{reason}}',
    'app.removeBlocked': 'Per rimuovere {{name}}, rimuovi prima: {{names}}.',
    'app.removeDataConfirm':
      'Rimuovere circa {{size}} di configurazione di {{name}}? La libreria e i download non verranno modificati. La configurazione sarà spostata nel cestino di Corsarr e potrà essere recuperata manualmente.',
    'app.dataArchived': 'La configurazione di {{name}} è stata spostata nel cestino di Corsarr.',
    'app.noData': '{{name}} non ha configurazioni da rimuovere.',
    'app.removeDataError':
      'Impossibile rimuovere la configurazione di {{name}}. Verifica che l’applicazione sia già stata rimossa.',
    'app.copyPasswordError': 'Impossibile copiare la password di {{name}}.',
    'app.addressCopied':
      'Indirizzo copiato. Aprilo su un dispositivo connesso alla stessa rete locale.',
    'app.addressCopyError':
      'Impossibile copiare l’indirizzo. Verifica che l’accesso dalla rete locale sia ancora attivo.',
    'storage.saved': 'Cartella salvata',
    'storage.savedDescription':
      'Corsarr verificherà di nuovo questa cartella prima di preparare le applicazioni.',
    'storage.savedBadge': 'Salvata',
    'storage.changeFolder': 'Cambia cartella',
    'setup.loadError': 'Impossibile ripristinare la preparazione salvata su questo computer.',
    'environment.readyTitle': 'Ambiente pronto',
    'environment.readyDescription': 'Questo computer è pronto per le applicazioni Corsarr.',
    'environment.readyBadge': 'Pronto',
    'environment.requiredTitle': 'Preparazione necessaria',
    'environment.requiredDescription':
      'Corsarr può preparare i componenti necessari dopo la tua autorizzazione esplicita.',
    'environment.requiredBadge': 'Non preparato',
    'environment.stoppedTitle': 'Ambiente arrestato',
    'environment.stoppedDescription':
      'Il componente necessario è installato, ma deve essere avviato.',
    'environment.attentionTitle': 'Impossibile verificare',
    'environment.attentionDescription':
      'Riprova. I dettagli tecnici possono aiutare nella diagnosi.',
    'environment.attentionBadge': 'Attenzione',
    'environment.start': 'Avvia ambiente',
    'environment.version': 'Versione: {{version}}',
    'environment.diagnostic': 'Diagnostica: {{detail}}',
    'environment.memory': 'Memoria: {{amount}} GiB',
    'environment.disk': 'Spazio libero: {{amount}} GiB',
    'environment.stoppedBadge': 'Arrestato',
    'environment.hostAttention': 'Questo computer richiede attenzione',
    'environment.requirements': 'Requisiti',
    'environment.provider': 'Provider: {{provider}}',
    'environment.state': 'Stato: {{state}}',
    'environment.retry': 'Riprova tra qualche istante.',
    'runtime.resolveRequirements':
      'Risolvi i requisiti indicati per questo computer prima di preparare l’ambiente.',
    'runtime.authorizationUnavailable':
      'L’autorizzazione iniziale non è disponibile. Riavvia la configurazione per preparare questo computer.',
    'runtime.installConfirm':
      'Corsarr scaricherà Docker Desktop 4.86.0 direttamente da Docker, ne verificherà il checksum e la firma ufficiale e richiederà l’autorizzazione di macOS per installarlo. Continuare?',
    'runtime.installing': 'Installazione…',
    'runtime.starting': 'Avvio…',
    'runtime.downloading':
      'Download e verifica dei componenti ufficiali. macOS potrebbe richiedere la password.',
    'runtime.startingDescription':
      'Avvio dei componenti necessari. Potrebbero essere necessari alcuni istanti.',
    'runtime.installedReady':
      'Questo computer è stato preparato ed è pronto per installare le applicazioni.',
    'runtime.startedReady': 'L’ambiente è stato avviato ed è pronto.',
    'runtime.incomplete':
      'La preparazione non è terminata. Non è stato installato nulla senza una firma valida; riprova o consulta i dettagli tecnici.',
    'runtime.error': 'Impossibile completare la preparazione di questo computer.',
  },
} as const;
