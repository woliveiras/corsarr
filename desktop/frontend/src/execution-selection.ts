import { GetExecutionConfiguration, SaveExecutionConfiguration } from '../wailsjs/go/main/App';
import { currentLocale } from './i18n';

type Configuration = { kind: string; context?: string };
let configuration: Configuration = { kind: 'docker-desktop' };
let desktopCopy: Map<Element, string> | undefined;
let renderedKey = '';

const messages = {
  en: {
    label: 'Container environment',
    existing: 'Use an existing Docker Engine',
    engine: 'Install Docker Engine on Linux',
    context: 'Local Docker context',
    save: 'Use this environment',
    description:
      'Choose how Corsarr will run your applications. Existing environments are managed by you or your administrator.',
    consent:
      'I authorize the selected environment and applications. Each component keeps its own license.',
    terms: 'Docker Engine license',
    install:
      'Install Docker Engine and Compose from Docker’s signed APT repository and enable the Docker service? Administrator authorization is required. Account permissions will not be changed.',
    existingHint:
      'Start the selected environment externally, then check again. Corsarr will not install Docker Desktop.',
    linuxHint:
      'Ubuntu 22.04/24.04/26.04 or Debian 12/13 with systemd. Your administrator may need to grant Docker socket access after installation.',
    failure:
      'Could not save the environment. Configure it before selecting storage or applications.',
  },
  'pt-BR': {
    label: 'Ambiente dos aplicativos',
    existing: 'Usar Docker Engine existente',
    engine: 'Instalar Docker Engine no Linux',
    context: 'Contexto Docker local',
    save: 'Usar este ambiente',
    description:
      'Escolha como o Corsarr executará seus aplicativos. Ambientes existentes são administrados por você ou pela TI.',
    consent:
      'Autorizo o ambiente selecionado e os aplicativos. Cada componente mantém sua própria licença.',
    terms: 'Licença do Docker Engine',
    install:
      'Instalar Docker Engine e Compose pelo repositório APT assinado da Docker e habilitar o serviço Docker? Será necessária autorização de administrador. As permissões da conta não serão alteradas.',
    existingHint:
      'Inicie o ambiente selecionado externamente e verifique novamente. O Corsarr não instalará Docker Desktop.',
    linuxHint:
      'Ubuntu 22.04/24.04/26.04 ou Debian 12/13 com systemd. A TI poderá precisar conceder acesso ao socket Docker após a instalação.',
    failure:
      'Não foi possível salvar o ambiente. Configure antes de selecionar armazenamento ou aplicativos.',
  },
  es: {
    label: 'Entorno de las aplicaciones',
    existing: 'Usar Docker Engine existente',
    engine: 'Instalar Docker Engine en Linux',
    context: 'Contexto Docker local',
    save: 'Usar este entorno',
    description:
      'Elige cómo Corsarr ejecutará tus aplicaciones. Tú o tu administrador gestionan los entornos existentes.',
    consent:
      'Autorizo el entorno seleccionado y las aplicaciones. Cada componente conserva su licencia.',
    terms: 'Licencia de Docker Engine',
    install:
      '¿Instalar Docker Engine y Compose desde el repositorio APT firmado de Docker y habilitar el servicio? Se requiere autorización de administrador. No se modificarán los permisos de la cuenta.',
    existingHint:
      'Inicia el entorno seleccionado externamente y comprueba de nuevo. Corsarr no instalará Docker Desktop.',
    linuxHint:
      'Ubuntu 22.04/24.04/26.04 o Debian 12/13 con systemd. Tu administrador podría necesitar conceder acceso al socket Docker.',
    failure:
      'No se pudo guardar el entorno. Configúralo antes de seleccionar almacenamiento o aplicaciones.',
  },
  it: {
    label: 'Ambiente delle applicazioni',
    existing: 'Usa Docker Engine esistente',
    engine: 'Installa Docker Engine su Linux',
    context: 'Contesto Docker locale',
    save: 'Usa questo ambiente',
    description:
      'Scegli come Corsarr eseguirà le applicazioni. Gli ambienti esistenti sono gestiti da te o dal tuo amministratore.',
    consent:
      'Autorizzo l’ambiente selezionato e le applicazioni. Ogni componente mantiene la propria licenza.',
    terms: 'Licenza Docker Engine',
    install:
      'Installare Docker Engine e Compose dal repository APT firmato di Docker e abilitare il servizio? È richiesta l’autorizzazione dell’amministratore. I permessi dell’account non saranno modificati.',
    existingHint:
      'Avvia esternamente l’ambiente selezionato e verifica di nuovo. Corsarr non installerà Docker Desktop.',
    linuxHint:
      'Ubuntu 22.04/24.04/26.04 o Debian 12/13 con systemd. L’amministratore potrebbe dover concedere l’accesso al socket Docker.',
    failure:
      'Impossibile salvare l’ambiente. Configuralo prima di selezionare archiviazione o applicazioni.',
  },
};

export function executionConfirmation(desktopMessage: string): string {
  if (configuration.kind === 'docker-engine') return messages[currentLocale()].install;
  if (configuration.kind === 'existing') return messages[currentLocale()].existingHint;
  return desktopMessage;
}

export function executionLegalID(): string {
  return configuration.kind === 'docker-desktop' ? 'runtime-docker' : 'runtime-docker-engine';
}

export function executionRequiresConfirmation(state: string): boolean {
  return (
    configuration.kind !== 'existing' &&
    (state === 'unavailable' || (configuration.kind === 'docker-engine' && state === 'stopped'))
  );
}

export function executionDescription(state: string, fallback: string): string {
  if (state === 'ready' || configuration.kind === 'docker-desktop') return fallback;
  const copy = messages[currentLocale()];
  return configuration.kind === 'existing' ? copy.existingHint : copy.linuxHint;
}

export async function loadExecutionSelection(
  platform: string,
  refresh: () => Promise<void>,
): Promise<void> {
  const loaded = await GetExecutionConfiguration();
  if (!loaded?.kind) return;
  configuration = loaded;
  const key = JSON.stringify([configuration, platform, currentLocale()]);
  if (renderedKey === key && document.querySelector('#execution-selection')) return;
  renderedKey = key;
  const copy = messages[currentLocale()];
  const parent = document.querySelector('#onboarding-permissions .onboarding-step-copy');
  if (!parent) return;
  if (!desktopCopy) {
    desktopCopy = new Map();
    for (const selector of [
      'p:not(.eyebrow)',
      'li',
      '#onboarding-terms + span',
      '#onboarding-open-docker-terms',
    ]) {
      const element = parent.querySelector(selector);
      if (element) desktopCopy.set(element, element.textContent ?? '');
    }
  }
  for (const [element, text] of desktopCopy) element.textContent = text;
  let panel = document.querySelector<HTMLElement>('#execution-selection');
  if (!panel) {
    panel = document.createElement('div');
    panel.id = 'execution-selection';
    panel.className = 'onboarding-explanation execution-selection';
    parent.querySelector('.onboarding-explanation')?.before(panel);
  }
  panel.replaceChildren();
  const label = document.createElement('label');
  label.textContent = copy.label;
  const select = document.createElement('select');
  select.id = 'execution-kind';
  const choices = [
    ['docker-desktop', 'Docker Desktop'],
    ['existing', copy.existing],
  ];
  if (platform === 'linux') choices.push(['docker-engine', copy.engine]);
  for (const [value, text] of choices) {
    const option = document.createElement('option');
    option.value = value;
    option.textContent = text;
    select.append(option);
  }
  select.value = configuration.kind;
  label.append(select);
  const contextLabel = document.createElement('label');
  contextLabel.textContent = copy.context;
  const context = document.createElement('input');
  context.value = configuration.context ?? 'default';
  context.id = 'execution-context';
  contextLabel.append(context);
  contextLabel.hidden = select.value !== 'existing';
  select.addEventListener('change', () => {
    contextLabel.hidden = select.value !== 'existing';
  });
  const termsCheckbox = document.querySelector<HTMLInputElement>('#onboarding-terms');
  const nextButton = document.querySelector<HTMLButtonElement>('#onboarding-permissions-next');
  const markDirty = () => {
    if (termsCheckbox) {
      termsCheckbox.checked = false;
      termsCheckbox.disabled = true;
    }
    if (nextButton) nextButton.disabled = true;
  };
  select.addEventListener('change', markDirty);
  context.addEventListener('input', markDirty);
  const button = document.createElement('button');
  button.type = 'button';
  button.textContent = copy.save;
  const hint = document.createElement('p');
  hint.role = 'status';
  hint.textContent = configuration.kind === 'docker-engine' ? copy.linuxHint : copy.existingHint;
  hint.hidden = configuration.kind === 'docker-desktop';
  button.addEventListener('click', async () => {
    button.disabled = true;
    try {
      configuration = await SaveExecutionConfiguration({
        kind: select.value,
        context: select.value === 'existing' ? context.value.trim() : '',
      });
      if (termsCheckbox) termsCheckbox.disabled = false;
      await refresh();
    } catch {
      hint.hidden = false;
      hint.textContent = copy.failure;
    } finally {
      button.disabled = false;
    }
  });
  panel.append(label, contextLabel, button, hint);
  // Keep authorization accurate for Engine without altering Desktop's wording.
  if (configuration.kind !== 'docker-desktop') {
    const description = parent.querySelector('p:not(.eyebrow)');
    if (description) description.textContent = copy.description;
    const authorized = parent.querySelector('li');
    if (authorized)
      authorized.textContent = configuration.kind === 'existing' ? copy.existing : copy.engine;
    const consent = parent.querySelector('#onboarding-terms + span');
    if (consent) consent.textContent = copy.consent;
    const terms = parent.querySelector('#onboarding-open-docker-terms');
    if (terms) terms.textContent = copy.terms;
  }
}
