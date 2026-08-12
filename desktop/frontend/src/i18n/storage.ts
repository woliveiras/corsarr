const en = {
  'storage.unknownSpace': 'Available space could not be determined',
  'storage.available': '{{amount}} GB available',
  'storage.checking': 'Checking…',
  'storage.readyDescription':
    'The folder is writable and supports efficient organization without duplicating files.',
  'storage.compatibleDescription':
    'The folder is writable but does not support hardlinks. Some imports may copy files.',
  'storage.hardlinks': 'Hardlinks available',
  'storage.noHardlinks': 'No hardlinks',
  'storage.ready': 'Ready',
  'storage.compatible': 'Compatible',
  'storage.folderReady': 'Folder ready',
  'storage.onboardingReady': 'The folder is writable and can organize files without duplication.',
  'storage.onboardingCompatible': 'The folder is writable; some imports may copy files.',
  'storage.cannotUse': 'This folder cannot be used',
  'storage.insufficient': 'Only {{available}} is available. Corsarr needs at least {{required}}.',
  'storage.chooseWritable':
    'Choose an existing folder with write permission and verifiable available space.',
  'storage.chooseAnotherWritable':
    'Choose another folder with write permission and available space.',
  'storage.verifyError': 'Could not check the folder',
  'storage.layoutReady':
    'Folder structure ready at {{path}}. No application has been installed yet.',
  'storage.layoutError':
    'Could not create the folders. Your selection was preserved for another attempt.',
  'storage.rechecking': 'Checking the folder again…',
  'storage.noLongerAvailable':
    'The folder is no longer available or has insufficient space. Choose another folder.',
} as const;
type StorageCatalog = Record<keyof typeof en, string>;
const es: StorageCatalog = {
  'storage.unknownSpace': 'No se pudo determinar el espacio disponible',
  'storage.available': '{{amount}} GB disponibles',
  'storage.checking': 'Comprobando…',
  'storage.readyDescription':
    'La carpeta permite escritura y organización eficiente sin duplicar archivos.',
  'storage.compatibleDescription':
    'La carpeta permite escritura, pero no admite hardlinks. Algunas importaciones pueden copiar archivos.',
  'storage.hardlinks': 'Hardlinks disponibles',
  'storage.noHardlinks': 'Sin hardlinks',
  'storage.ready': 'Lista',
  'storage.compatible': 'Compatible',
  'storage.folderReady': 'Carpeta lista',
  'storage.onboardingReady': 'La carpeta permite escritura y organizar archivos sin duplicación.',
  'storage.onboardingCompatible':
    'La carpeta permite escritura; algunas importaciones pueden copiar archivos.',
  'storage.cannotUse': 'Esta carpeta no se puede usar',
  'storage.insufficient': 'Solo hay {{available}}. Corsarr necesita al menos {{required}}.',
  'storage.chooseWritable':
    'Elige una carpeta existente con permiso de escritura y espacio disponible verificable.',
  'storage.chooseAnotherWritable':
    'Elige otra carpeta con permiso de escritura y espacio disponible.',
  'storage.verifyError': 'No se pudo comprobar la carpeta',
  'storage.layoutReady': 'Estructura lista en {{path}}. Aún no se instaló ninguna aplicación.',
  'storage.layoutError':
    'No se pudieron crear las carpetas. Se conservó tu selección para otro intento.',
  'storage.rechecking': 'Comprobando la carpeta de nuevo…',
  'storage.noLongerAvailable':
    'La carpeta ya no está disponible o no tiene espacio suficiente. Elige otra carpeta.',
};
const ptBR: StorageCatalog = {
  'storage.unknownSpace': 'Espaço disponível não identificado',
  'storage.available': '{{amount}} GB disponíveis',
  'storage.checking': 'Verificando…',
  'storage.readyDescription':
    'A pasta é gravável e suporta organização eficiente sem duplicar arquivos.',
  'storage.compatibleDescription':
    'A pasta é gravável, mas não oferece hardlinks. Algumas importações poderão copiar arquivos.',
  'storage.hardlinks': 'Hardlinks disponíveis',
  'storage.noHardlinks': 'Sem hardlinks',
  'storage.ready': 'Pronto',
  'storage.compatible': 'Compatível',
  'storage.folderReady': 'Pasta pronta',
  'storage.onboardingReady': 'A pasta é gravável e permite organizar arquivos sem duplicação.',
  'storage.onboardingCompatible':
    'A pasta é gravável; algumas importações poderão copiar arquivos.',
  'storage.cannotUse': 'Esta pasta não pode ser usada',
  'storage.insufficient': 'Há apenas {{available}}. O Corsarr precisa de pelo menos {{required}}.',
  'storage.chooseWritable':
    'Escolha uma pasta existente com permissão de escrita e espaço disponível verificável.',
  'storage.chooseAnotherWritable':
    'Escolha outra pasta com permissão de escrita e espaço disponível.',
  'storage.verifyError': 'Não foi possível verificar a pasta',
  'storage.layoutReady': 'Estrutura pronta em {{path}}. Nenhum aplicativo foi instalado ainda.',
  'storage.layoutError':
    'Não foi possível criar as pastas. Sua seleção foi preservada para uma nova tentativa.',
  'storage.rechecking': 'Verificando a pasta novamente…',
  'storage.noLongerAvailable':
    'A pasta não está mais disponível ou não possui espaço suficiente. Escolha outra pasta.',
};
const it: StorageCatalog = {
  'storage.unknownSpace': 'Impossibile determinare lo spazio disponibile',
  'storage.available': '{{amount}} GB disponibili',
  'storage.checking': 'Verifica…',
  'storage.readyDescription':
    'La cartella è scrivibile e consente un’organizzazione efficiente senza duplicare file.',
  'storage.compatibleDescription':
    'La cartella è scrivibile ma non supporta gli hardlink. Alcune importazioni potrebbero copiare i file.',
  'storage.hardlinks': 'Hardlink disponibili',
  'storage.noHardlinks': 'Nessun hardlink',
  'storage.ready': 'Pronta',
  'storage.compatible': 'Compatibile',
  'storage.folderReady': 'Cartella pronta',
  'storage.onboardingReady':
    'La cartella è scrivibile e consente di organizzare i file senza duplicarli.',
  'storage.onboardingCompatible':
    'La cartella è scrivibile; alcune importazioni potrebbero copiare i file.',
  'storage.cannotUse': 'Questa cartella non può essere usata',
  'storage.insufficient':
    'Sono disponibili solo {{available}}. Corsarr richiede almeno {{required}}.',
  'storage.chooseWritable':
    'Scegli una cartella esistente con permesso di scrittura e spazio disponibile verificabile.',
  'storage.chooseAnotherWritable':
    'Scegli un’altra cartella con permesso di scrittura e spazio disponibile.',
  'storage.verifyError': 'Impossibile verificare la cartella',
  'storage.layoutReady':
    'Struttura pronta in {{path}}. Nessuna applicazione è stata ancora installata.',
  'storage.layoutError':
    'Impossibile creare le cartelle. La selezione è stata conservata per un nuovo tentativo.',
  'storage.rechecking': 'Nuova verifica della cartella…',
  'storage.noLongerAvailable':
    'La cartella non è più disponibile o non ha spazio sufficiente. Scegli un’altra cartella.',
};
export const storageMessages = { en, es, 'pt-BR': ptBR, it } as const;
