import './style.css';
import {
  ChooseStorageLocation,
  GetEnvironmentStatus,
  ListApplications,
  OpenApplication,
} from '../wailsjs/go/main/App';
import type { application } from '../wailsjs/go/models';

type Application = application.ApplicationSummary;

const root = document.querySelector<HTMLDivElement>('#app');

if (!root) {
  throw new Error('Elemento principal do Corsarr não encontrado.');
}

root.innerHTML = [
  '<div class="shell">',
  '  <aside class="sidebar">',
  '    <a class="brand" href="#" aria-label="Corsarr, início">',
  '      <span class="brand-mark" aria-hidden="true">C</span>',
  '      <span><strong>Corsarr</strong><small>Desktop preview</small></span>',
  '    </a>',
  '    <nav aria-label="Navegação principal">',
  '      <button class="nav-item active" type="button"><span aria-hidden="true">⌂</span>Início</button>',
  '      <button class="nav-item" type="button" disabled><span aria-hidden="true">⊞</span>Aplicativos<small>em breve</small></button>',
  '      <button class="nav-item" type="button" disabled><span aria-hidden="true">⚙</span>Ajustes<small>em breve</small></button>',
  '    </nav>',
  '    <div class="sidebar-note">',
  '      <span class="status-dot"></span>',
  '      <p><strong>Primeiro marco</strong>Interface conectada ao catálogo real do Corsarr.</p>',
  '    </div>',
  '  </aside>',
  '  <main class="content">',
  '    <header class="topbar">',
  '      <div><p class="eyebrow">SEU SERVIDOR DE MÍDIA</p><h1>Olá, William.</h1></div>',
  '      <div class="machine"><span class="machine-icon" aria-hidden="true">⌘</span><span><small id="platform-name">Verificando computador</small><strong id="architecture-name">Aguarde…</strong></span></div>',
  '    </header>',
  '    <section class="hero" aria-labelledby="hero-title">',
  '      <div>',
  '        <span class="phase">FASE 1 · FUNDAÇÃO</span>',
  '        <h2 id="hero-title">Tudo no lugar certo,<br><em>sem complicação.</em></h2>',
  '        <p>O Corsarr está preparando uma experiência única para instalar e cuidar dos seus aplicativos de mídia.</p>',
  '      </div>',
  '      <div class="radar" aria-hidden="true"><span></span><span></span><i></i><b>C</b></div>',
  '    </section>',
  '    <section class="environment-card" aria-labelledby="environment-title">',
  '      <span class="environment-icon" aria-hidden="true">◎</span>',
  '      <div class="environment-copy">',
  '        <p class="eyebrow">ESTE COMPUTADOR</p>',
  '        <h2 id="environment-title">Verificando o ambiente…</h2>',
  '        <p id="environment-description">O Corsarr está conferindo os componentes necessários.</p>',
  '        <details id="environment-details"><summary>Detalhes técnicos</summary><code id="environment-technical">Executando diagnóstico…</code></details>',
  '      </div>',
  '      <span id="environment-badge" class="runtime-badge checking">Verificando</span>',
  '      <button id="refresh-environment" class="refresh-button" type="button" aria-label="Verificar ambiente novamente">↻</button>',
  '    </section>',
  '    <section class="storage-card" aria-labelledby="storage-title">',
  '      <span class="storage-icon" aria-hidden="true">▱</span>',
  '      <div class="storage-copy">',
  '        <p class="eyebrow">ARMAZENAMENTO</p>',
  '        <h2 id="storage-title">Escolha onde guardar sua mídia</h2>',
  '        <p id="storage-description">O Corsarr verificará espaço, permissão de escrita e suporte a hardlinks.</p>',
  '        <p id="storage-path" class="storage-path"></p>',
  '        <p id="storage-facts" class="storage-facts"></p>',
  '      </div>',
  '      <span id="storage-badge" class="runtime-badge checking">Não verificado</span>',
  '      <button id="choose-storage" class="choose-storage-button" type="button">Escolher pasta</button>',
  '    </section>',
  '    <section class="section-heading">',
  '      <div><p class="eyebrow">CATÁLOGO DETECTADO</p><h2>Seus aplicativos</h2></div>',
  '      <p id="catalog-count" class="catalog-count">Carregando…</p>',
  '    </section>',
  '    <div id="message" class="message" role="status" aria-live="polite"></div>',
  '    <section id="applications" class="applications" aria-label="Aplicativos disponíveis">',
  '      <div class="loading-card"></div><div class="loading-card"></div><div class="loading-card"></div>',
  '    </section>',
  '  </main>',
  '</div>',
].join('');

const applicationsElement = document.querySelector<HTMLElement>('#applications');
const countElement = document.querySelector<HTMLElement>('#catalog-count');
const messageElement = document.querySelector<HTMLElement>('#message');
const platformElement = document.querySelector<HTMLElement>('#platform-name');
const architectureElement = document.querySelector<HTMLElement>('#architecture-name');
const environmentTitleElement = document.querySelector<HTMLElement>('#environment-title');
const environmentDescriptionElement = document.querySelector<HTMLElement>(
  '#environment-description',
);
const environmentTechnicalElement = document.querySelector<HTMLElement>('#environment-technical');
const environmentBadgeElement = document.querySelector<HTMLElement>('#environment-badge');
const refreshEnvironmentButton = document.querySelector<HTMLButtonElement>('#refresh-environment');
const storageTitleElement = document.querySelector<HTMLElement>('#storage-title');
const storageDescriptionElement = document.querySelector<HTMLElement>('#storage-description');
const storagePathElement = document.querySelector<HTMLElement>('#storage-path');
const storageFactsElement = document.querySelector<HTMLElement>('#storage-facts');
const storageBadgeElement = document.querySelector<HTMLElement>('#storage-badge');
const chooseStorageButton = document.querySelector<HTMLButtonElement>('#choose-storage');

const symbols: Record<string, string> = {
  bazarr: 'Bz',
  fileflows: 'Ff',
  jellyfin: 'Jf',
  jellyseerr: 'Js',
  lazylibrarian: 'Ll',
  lidarr: 'Li',
  prowlarr: 'Pr',
  qbittorrent: 'qB',
  radarr: 'Ra',
  sonarr: 'So',
};

function createApplicationCard(application: Application): HTMLElement {
  const card = document.createElement('article');
  card.className = 'application-card';

  const icon = document.createElement('div');
  icon.className = `application-icon ${application.id}`;
  icon.textContent = symbols[application.id] ?? application.name.slice(0, 2);
  icon.setAttribute('aria-hidden', 'true');

  const information = document.createElement('div');
  information.className = 'application-info';

  const title = document.createElement('h3');
  title.textContent = application.name;

  const description = document.createElement('p');
  description.textContent = application.description;

  const metadata = document.createElement('div');
  metadata.className = 'metadata';
  metadata.textContent = application.optional ? 'Opcional' : 'Aplicativo principal';

  information.append(title, description, metadata);

  const button = document.createElement('button');
  button.className = 'open-button';
  button.type = 'button';
  button.textContent = 'Abrir';
  button.setAttribute('aria-label', `Abrir ${application.name} no navegador`);
  button.addEventListener('click', async () => {
    button.disabled = true;
    messageElement?.classList.remove('error');
    if (messageElement) messageElement.textContent = `Abrindo ${application.name}…`;

    try {
      await OpenApplication(application.id);
      if (messageElement)
        messageElement.textContent = `${application.name} foi aberto no navegador.`;
    } catch {
      if (messageElement) {
        messageElement.textContent = `Não foi possível abrir ${application.name}. O serviço ainda pode não estar instalado.`;
        messageElement.classList.add('error');
      }
    } finally {
      button.disabled = false;
    }
  });

  card.append(icon, information, button);
  return card;
}

async function loadApplications(): Promise<void> {
  if (!applicationsElement || !countElement) return;

  try {
    const applications = await ListApplications();
    applicationsElement.replaceChildren(...applications.map(createApplicationCard));
    countElement.textContent = `${applications.length} disponíveis`;
  } catch {
    applicationsElement.replaceChildren();
    countElement.textContent = 'Indisponível';
    if (messageElement) {
      messageElement.textContent = 'Não foi possível carregar o catálogo do Corsarr.';
      messageElement.classList.add('error');
    }
  }
}

const platformNames: Record<string, string> = {
  darwin: 'macOS',
  linux: 'Linux',
  windows: 'Windows',
};

const architectureNames: Record<string, string> = {
  arm64: 'Apple Silicon / ARM64',
  amd64: 'Intel / AMD64',
};

const runtimeMessages: Record<string, { title: string; description: string; badge: string }> = {
  ready: {
    title: 'Ambiente pronto',
    description: 'Este computador está pronto para receber os aplicativos do Corsarr.',
    badge: 'Pronto',
  },
  unavailable: {
    title: 'Preparação necessária',
    description:
      'Os componentes necessários ainda não estão instalados. A instalação guiada será habilitada no próximo marco.',
    badge: 'Não preparado',
  },
  stopped: {
    title: 'Ambiente pausado',
    description: 'O componente necessário está instalado, mas precisa ser iniciado.',
    badge: 'Parado',
  },
  error: {
    title: 'Não foi possível verificar',
    description: 'Tente novamente. Os detalhes técnicos podem ajudar no diagnóstico.',
    badge: 'Atenção',
  },
};

async function loadEnvironment(): Promise<void> {
  refreshEnvironmentButton?.setAttribute('disabled', 'true');

  try {
    const environment = await GetEnvironmentStatus();
    const runtimeMessage = runtimeMessages[environment.runtime.state] ?? runtimeMessages.error;

    if (platformElement) {
      platformElement.textContent = platformNames[environment.platform] ?? environment.platform;
    }
    if (architectureElement) {
      architectureElement.textContent =
        architectureNames[environment.architecture] ?? environment.architecture;
    }
    if (environmentTitleElement) environmentTitleElement.textContent = runtimeMessage.title;
    if (environmentDescriptionElement) {
      environmentDescriptionElement.textContent = runtimeMessage.description;
    }
    if (environmentBadgeElement) {
      environmentBadgeElement.textContent = runtimeMessage.badge;
      environmentBadgeElement.className = `runtime-badge ${environment.runtime.state}`;
    }
    if (environmentTechnicalElement) {
      const details = [
        `Provedor: ${environment.runtime.provider}`,
        `Estado: ${environment.runtime.state}`,
      ];
      if (environment.runtime.version) details.push(`Versão: ${environment.runtime.version}`);
      if (environment.runtime.technicalDetail) {
        details.push(`Diagnóstico: ${environment.runtime.technicalDetail}`);
      }
      environmentTechnicalElement.textContent = details.join('\n');
    }
  } catch {
    if (environmentTitleElement) environmentTitleElement.textContent = 'Não foi possível verificar';
    if (environmentDescriptionElement) {
      environmentDescriptionElement.textContent = 'Tente novamente em alguns instantes.';
    }
    if (environmentBadgeElement) {
      environmentBadgeElement.textContent = 'Atenção';
      environmentBadgeElement.className = 'runtime-badge error';
    }
  } finally {
    refreshEnvironmentButton?.removeAttribute('disabled');
  }
}

refreshEnvironmentButton?.addEventListener('click', () => void loadEnvironment());

function formatAvailableSpace(bytes: number): string {
  if (bytes <= 0) return 'Espaço disponível não identificado';
  return `${new Intl.NumberFormat('pt-BR', { maximumFractionDigits: 1 }).format(
    bytes / 1024 ** 3,
  )} GB disponíveis`;
}

async function chooseStorage(): Promise<void> {
  if (!chooseStorageButton) return;
  chooseStorageButton.disabled = true;
  chooseStorageButton.textContent = 'Verificando…';

  try {
    const storage = await ChooseStorageLocation();
    if (storage.state === 'canceled') return;

    if (storage.state === 'ready') {
      if (storageTitleElement) storageTitleElement.textContent = 'Armazenamento pronto';
      if (storageDescriptionElement) {
        storageDescriptionElement.textContent = storage.hardlinks
          ? 'A pasta é gravável e suporta organização eficiente sem duplicar arquivos.'
          : 'A pasta é gravável, mas não oferece hardlinks. Algumas importações poderão copiar arquivos.';
      }
      if (storagePathElement) storagePathElement.textContent = storage.path;
      if (storageFactsElement) {
        storageFactsElement.textContent = `${formatAvailableSpace(storage.availableBytes ?? 0)} · ${
          storage.hardlinks ? 'Hardlinks disponíveis' : 'Sem hardlinks'
        }`;
      }
      if (storageBadgeElement) {
        storageBadgeElement.textContent = storage.hardlinks ? 'Pronto' : 'Compatível';
        storageBadgeElement.className = `runtime-badge ${storage.hardlinks ? 'ready' : 'stopped'}`;
      }
      chooseStorageButton.textContent = 'Trocar pasta';
      return;
    }

    if (storageTitleElement) storageTitleElement.textContent = 'Esta pasta não pode ser usada';
    if (storageDescriptionElement) {
      storageDescriptionElement.textContent =
        storage.technicalDetail ?? 'Escolha uma pasta existente com permissão de escrita.';
    }
    if (storagePathElement) storagePathElement.textContent = storage.path;
    if (storageFactsElement) storageFactsElement.textContent = '';
    if (storageBadgeElement) {
      storageBadgeElement.textContent = 'Atenção';
      storageBadgeElement.className = 'runtime-badge error';
    }
  } catch {
    if (storageTitleElement) storageTitleElement.textContent = 'Não foi possível verificar a pasta';
    if (storageDescriptionElement) storageDescriptionElement.textContent = 'Tente novamente.';
    if (storageBadgeElement) {
      storageBadgeElement.textContent = 'Atenção';
      storageBadgeElement.className = 'runtime-badge error';
    }
  } finally {
    chooseStorageButton.disabled = false;
    if (chooseStorageButton.textContent === 'Verificando…') {
      chooseStorageButton.textContent = 'Escolher pasta';
    }
  }
}

chooseStorageButton?.addEventListener('click', () => void chooseStorage());

void loadApplications();
void loadEnvironment();
