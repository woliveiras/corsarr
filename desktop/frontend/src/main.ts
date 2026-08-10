import './style.css';
import {
  AcceptCurrentTerms,
  ArchiveApplicationData,
  ChooseStorageLocation,
  CopyJellyfinPassword,
  CopyQBittorrentPassword,
  ExportDiagnostics,
  GetApplicationDataStatuses,
  GetApplicationStatuses,
  GetEnvironmentStatus,
  GetJellyfinAccessStatus,
  GetQBittorrentAccessStatus,
  GetSetupStatus,
  InstallSelectedApplications,
  ListApplications,
  ListLegalNotices,
  OpenApplication,
  OpenLegalLink,
  PrepareRuntime,
  PrepareStorageLayout,
  RemoveApplication,
  RestartApplication,
  SaveApplicationSelection,
  StartApplication,
  StopApplication,
  UpdateApplication,
} from '../wailsjs/go/main/App';
import type { application, legal, storage } from '../wailsjs/go/models';

type Application = application.ApplicationSummary;
type ManagedStatus = application.ManagedApplicationStatus;
type DataStatus = storage.ApplicationDataStatus;
type LegalNotice = legal.Notice;

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
  '      <button id="show-home" class="nav-item active" type="button"><span aria-hidden="true">⌂</span>Início</button>',
  '      <button class="nav-item" type="button" disabled><span aria-hidden="true">⊞</span>Aplicativos<small>em breve</small></button>',
  '      <button id="show-licenses" class="nav-item" type="button"><span aria-hidden="true">§</span>Licenças</button>',
  '      <button id="export-diagnostics" class="nav-item" type="button"><span aria-hidden="true">⇩</span>Exportar diagnóstico</button>',
  '    </nav>',
  '    <div class="sidebar-note">',
  '      <span class="status-dot"></span>',
  '      <p><strong>Primeiro marco</strong>Interface conectada ao catálogo real do Corsarr.</p>',
  '    </div>',
  '  </aside>',
  '  <main class="content">',
  '    <div id="home-view">',
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
  '      <button id="prepare-runtime" class="choose-storage-button" type="button" hidden>Preparar computador</button>',
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
  '    <section class="installation-review" aria-labelledby="installation-title">',
  '      <div>',
  '        <p class="eyebrow">PRÓXIMA ETAPA</p>',
  '        <h2 id="installation-title">Revise sua preparação</h2>',
  '        <p id="installation-summary">Escolha uma pasta e ao menos um aplicativo.</p>',
  '        <label class="terms-consent"><input id="accept-terms" type="checkbox"> <span>Autorizo o Corsarr a instalar e usar o Docker Desktop, baixar as imagens aprovadas e criar os serviços selecionados. Aceito os termos do Docker Desktop e entendo que o uso pessoal é gratuito, enquanto empresas maiores e entidades governamentais podem precisar de assinatura. Cada aplicação mantém sua própria licença.</span></label>',
  '        <p id="installation-result" class="installation-result"></p>',
  '      </div>',
  '      <div class="installation-actions">',
  '        <button id="prepare-storage" class="secondary-button" type="button" disabled>Preparar pastas</button>',
  '        <button id="install-applications" class="prepare-button" type="button" disabled>Instalar aplicativos</button>',
  '      </div>',
  '    </section>',
  '    </div>',
  '    <section id="licenses-view" class="licenses-view" hidden aria-labelledby="licenses-title">',
  '      <header class="credits-header"><div><p class="eyebrow">CRÉDITOS E TRANSPARÊNCIA</p><h1 id="licenses-title">Aplicativos e licenças</h1><p>O Corsarr existe graças a estes projetos e mantenedores. Cada componente mantém seus próprios termos, marcas e direitos autorais.</p></div><button id="licenses-back" class="secondary-button" type="button">Voltar ao início</button></header>',
  '      <p class="affiliation-note">O Corsarr facilita a instalação e a administração, mas não é afiliado ou endossado pelos projetos listados, salvo indicação expressa.</p>',
  '      <div id="legal-notices" class="legal-notices"><div class="loading-card"></div></div>',
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
const prepareRuntimeButton = document.querySelector<HTMLButtonElement>('#prepare-runtime');
const storageTitleElement = document.querySelector<HTMLElement>('#storage-title');
const storageDescriptionElement = document.querySelector<HTMLElement>('#storage-description');
const storagePathElement = document.querySelector<HTMLElement>('#storage-path');
const storageFactsElement = document.querySelector<HTMLElement>('#storage-facts');
const storageBadgeElement = document.querySelector<HTMLElement>('#storage-badge');
const chooseStorageButton = document.querySelector<HTMLButtonElement>('#choose-storage');
const installationSummaryElement = document.querySelector<HTMLElement>('#installation-summary');
const installationResultElement = document.querySelector<HTMLElement>('#installation-result');
const prepareStorageButton = document.querySelector<HTMLButtonElement>('#prepare-storage');
const installApplicationsButton =
  document.querySelector<HTMLButtonElement>('#install-applications');
const acceptTermsCheckbox = document.querySelector<HTMLInputElement>('#accept-terms');
const homeView = document.querySelector<HTMLElement>('#home-view');
const licensesView = document.querySelector<HTMLElement>('#licenses-view');
const showHomeButton = document.querySelector<HTMLButtonElement>('#show-home');
const showLicensesButton = document.querySelector<HTMLButtonElement>('#show-licenses');
const exportDiagnosticsButton = document.querySelector<HTMLButtonElement>('#export-diagnostics');
const licensesBackButton = document.querySelector<HTMLButtonElement>('#licenses-back');
const legalNoticesElement = document.querySelector<HTMLElement>('#legal-notices');

let setupStatus: application.SetupStatus | undefined;
let availableApplications: Application[] = [];
let selectedApplicationIDs = new Set<string>();
let selectionSaving = false;
let managedStatuses = new Map<string, ManagedStatus>();
let dataStatuses = new Map<string, DataStatus>();
let qbittorrentAccess: application.ServiceAccessStatus | undefined;
let jellyfinAccess: application.ServiceAccessStatus | undefined;
let legalNotices: LegalNotice[] = [];
let currentRuntimeState = 'checking';

function showView(view: 'home' | 'licenses'): void {
  const showHome = view === 'home';
  if (homeView) homeView.hidden = !showHome;
  if (licensesView) licensesView.hidden = showHome;
  showHomeButton?.classList.toggle('active', showHome);
  showLicensesButton?.classList.toggle('active', !showHome);
  document.querySelector('.content')?.scrollTo({ top: 0, behavior: 'smooth' });
}

showHomeButton?.addEventListener('click', () => showView('home'));
showLicensesButton?.addEventListener('click', () => showView('licenses'));
licensesBackButton?.addEventListener('click', () => showView('home'));

exportDiagnosticsButton?.addEventListener('click', async () => {
  if (!exportDiagnosticsButton) return;
  exportDiagnosticsButton.disabled = true;
  try {
    const result = await ExportDiagnostics();
    if (!result.exported) return;
    if (messageElement) {
      messageElement.textContent = `Diagnóstico salvo em ${result.path}. O arquivo não inclui logs nem credenciais.`;
      messageElement.classList.remove('error');
    }
    showView('home');
  } catch {
    if (messageElement) {
      messageElement.textContent = 'Não foi possível exportar o diagnóstico.';
      messageElement.classList.add('error');
    }
    showView('home');
  } finally {
    exportDiagnosticsButton.disabled = false;
  }
});

const symbols: Record<string, string> = {
  bazarr: 'Bz',
  fileflows: 'Ff',
  jellyfin: 'Jf',
  jellyseerr: 'Sr',
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
  if (selectedApplicationIDs.has(application.id)) card.classList.add('selected');

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
  const managedStatus = managedStatuses.get(application.id);
  const stateLabels: Record<string, string> = {
    not_installed: 'Não instalado',
    running: 'Em execução',
    stopped: 'Parado',
    attention: 'Atenção',
  };
  metadata.textContent = `${application.optional ? 'Opcional' : 'Aplicativo principal'} · ${stateLabels[managedStatus?.state ?? 'not_installed']}${managedStatus?.updateAvailable ? ' · Atualização disponível' : ''}`;

  information.append(title, description, metadata);

  const actions = document.createElement('div');
  actions.className = 'application-actions';

  const selectButton = document.createElement('button');
  selectButton.className = 'select-button';
  selectButton.type = 'button';
  selectButton.textContent = selectedApplicationIDs.has(application.id)
    ? 'Selecionado'
    : 'Selecionar';
  selectButton.setAttribute('aria-pressed', String(selectedApplicationIDs.has(application.id)));
  selectButton.setAttribute('aria-label', `Selecionar ${application.name} para instalação`);
  selectButton.disabled = selectionSaving;
  selectButton.addEventListener('click', async () => {
    if (selectionSaving) return;
    const previousSelection = new Set(selectedApplicationIDs);
    if (selectedApplicationIDs.has(application.id)) {
      selectedApplicationIDs.delete(application.id);
    } else {
      selectedApplicationIDs.add(application.id);
    }
    selectionSaving = true;
    renderApplications();

    try {
      const status = await SaveApplicationSelection([...selectedApplicationIDs]);
      applySetupStatus(status);
      messageElement?.classList.remove('error');
    } catch {
      selectedApplicationIDs = previousSelection;
      if (messageElement) {
        messageElement.textContent = 'Não foi possível salvar sua seleção de aplicativos.';
        messageElement.classList.add('error');
      }
    } finally {
      selectionSaving = false;
      renderApplications();
    }
  });

  const openButton = document.createElement('button');
  openButton.className = 'open-button';
  openButton.type = 'button';
  openButton.textContent = 'Abrir';
  openButton.disabled = managedStatus?.state !== 'running';
  openButton.setAttribute('aria-label', `Abrir ${application.name} no navegador`);
  openButton.addEventListener('click', async () => {
    openButton.disabled = true;
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
      openButton.disabled = false;
    }
  });

  actions.append(selectButton, openButton);
  if (
    managedStatus?.updateAvailable &&
    (managedStatus.state === 'running' || managedStatus.state === 'stopped')
  ) {
    actions.append(updateApplicationButton(application));
  }
  if (managedStatus?.state === 'running') {
    actions.append(
      lifecycleButton('Reiniciar', () => RestartApplication(application.id)),
      lifecycleButton('Parar', () => StopApplication(application.id)),
      lifecycleButton('Remover', () => removeApplication(application)),
    );
  } else if (managedStatus?.state === 'stopped') {
    actions.append(
      lifecycleButton('Iniciar', () => StartApplication(application.id)),
      lifecycleButton('Remover', () => removeApplication(application)),
    );
  } else if (
    managedStatus?.state === 'not_installed' &&
    dataStatuses.get(application.id)?.present
  ) {
    actions.append(dataRemovalButton(application));
  }
  if (
    application.id === 'qbittorrent' &&
    qbittorrentAccess?.available &&
    (managedStatus?.state === 'running' || managedStatus?.state === 'stopped')
  ) {
    actions.append(qbittorrentCredentialButton());
  }
  if (
    application.id === 'jellyfin' &&
    jellyfinAccess?.available &&
    (managedStatus?.state === 'running' || managedStatus?.state === 'stopped')
  ) {
    actions.append(jellyfinCredentialButton());
  }
  card.append(icon, information, actions);
  return card;
}

function updateApplicationButton(target: Application): HTMLButtonElement {
  const button = document.createElement('button');
  button.className = 'update-button';
  button.type = 'button';
  button.textContent = 'Atualizar';
  button.addEventListener('click', async () => {
    const confirmed = window.confirm(
      `Atualizar ${target.name}? O Corsarr criará um backup privado das configurações e verificará a nova versão antes de concluir. O aplicativo ficará indisponível por alguns instantes. Se a versão migrar o banco de dados, restaurar a imagem anterior pode não desfazer essa migração.`,
    );
    if (!confirmed) return;

    button.disabled = true;
    button.textContent = 'Atualizando…';
    if (messageElement) {
      messageElement.textContent = `Criando backup e verificando a atualização de ${target.name}…`;
      messageElement.classList.remove('error');
    }
    try {
      const result = await UpdateApplication(target.id);
      if (messageElement) {
        if (result.updated && !result.error) {
          messageElement.textContent = `${target.name} foi atualizado e verificado. O backup das configurações foi preservado.`;
          messageElement.classList.remove('error');
        } else if (result.rolledBack) {
          messageElement.textContent = `A nova versão de ${target.name} não passou na verificação. A imagem anterior foi restaurada e o backup foi preservado.`;
          messageElement.classList.add('error');
        } else if (result.error) {
          messageElement.textContent = `${target.name} requer atenção após a tentativa de atualização. Consulte os detalhes técnicos.`;
          messageElement.classList.add('error');
        } else {
          messageElement.textContent = `${target.name} já usa a versão aprovada pelo Corsarr.`;
          messageElement.classList.remove('error');
        }
      }
      await loadApplicationStatuses();
    } catch {
      if (messageElement) {
        messageElement.textContent = `Não foi possível iniciar a atualização de ${target.name}. Nenhuma alteração foi autorizada fora dos recursos do Corsarr.`;
        messageElement.classList.add('error');
      }
    } finally {
      button.disabled = false;
      button.textContent = 'Atualizar';
    }
  });
  return button;
}

function lifecycleButton(label: string, operation: () => Promise<void>): HTMLButtonElement {
  const button = document.createElement('button');
  button.className = 'lifecycle-button';
  button.type = 'button';
  button.textContent = label;
  button.addEventListener('click', async () => {
    button.disabled = true;
    try {
      await operation();
      await loadApplicationStatuses();
    } catch {
      if (messageElement) {
        messageElement.textContent = `Não foi possível ${label.toLocaleLowerCase('pt-BR')} o aplicativo.`;
        messageElement.classList.add('error');
      }
    } finally {
      button.disabled = false;
    }
  });
  return button;
}

async function removeApplication(target: Application): Promise<void> {
  const confirmed = window.confirm(
    `Remover ${target.name}? As configurações e sua mídia serão preservadas.`,
  );
  if (!confirmed) return;
  await RemoveApplication(target.id);
}

function dataRemovalButton(target: Application): HTMLButtonElement {
  const button = document.createElement('button');
  button.className = 'data-removal-button';
  button.type = 'button';
  button.textContent = 'Remover dados';
  button.addEventListener('click', async () => {
    const confirmed = window.confirm(
      `Remover as configurações de ${target.name}? A biblioteca e os downloads não serão alterados. A configuração será movida para a lixeira do Corsarr.`,
    );
    if (!confirmed) return;

    button.disabled = true;
    try {
      const archived = await ArchiveApplicationData(target.id);
      if (messageElement) {
        messageElement.textContent = archived.archived
          ? `As configurações de ${target.name} foram movidas para a lixeira do Corsarr.`
          : `${target.name} não possui configurações para remover.`;
        messageElement.classList.remove('error');
      }
      await loadApplicationDataStatuses();
    } catch {
      if (messageElement) {
        messageElement.textContent = `Não foi possível remover as configurações de ${target.name}. Confirme que o aplicativo já foi removido.`;
        messageElement.classList.add('error');
      }
    } finally {
      button.disabled = false;
    }
  });
  return button;
}

function qbittorrentCredentialButton(): HTMLButtonElement {
  const button = document.createElement('button');
  button.className = 'credential-button';
  button.type = 'button';
  button.textContent = 'Copiar senha';
  button.addEventListener('click', async () => {
    button.disabled = true;
    try {
      await CopyQBittorrentPassword();
      if (messageElement) {
        messageElement.textContent = `Senha copiada. Use o usuário ${qbittorrentAccess?.username ?? 'corsarr'} para entrar no qBittorrent.`;
        messageElement.classList.remove('error');
      }
    } catch {
      if (messageElement) {
        messageElement.textContent = 'Não foi possível copiar a senha do qBittorrent.';
        messageElement.classList.add('error');
      }
    } finally {
      button.disabled = false;
    }
  });
  return button;
}

function jellyfinCredentialButton(): HTMLButtonElement {
  const button = document.createElement('button');
  button.className = 'credential-button';
  button.type = 'button';
  button.textContent = 'Copiar senha';
  button.addEventListener('click', async () => {
    button.disabled = true;
    try {
      await CopyJellyfinPassword();
      if (messageElement) {
        messageElement.textContent = `Senha copiada. Use o usuário ${jellyfinAccess?.username ?? 'corsarr'} para entrar no Jellyfin.`;
        messageElement.classList.remove('error');
      }
    } catch {
      if (messageElement) {
        messageElement.textContent = 'Não foi possível copiar a senha do Jellyfin.';
        messageElement.classList.add('error');
      }
    } finally {
      button.disabled = false;
    }
  });
  return button;
}

function renderApplications(): void {
  applicationsElement?.replaceChildren(...availableApplications.map(createApplicationCard));
}

async function loadApplications(): Promise<void> {
  if (!applicationsElement || !countElement) return;

  try {
    availableApplications = await ListApplications();
    renderApplications();
    countElement.textContent = `${availableApplications.length} disponíveis`;
  } catch {
    applicationsElement.replaceChildren();
    countElement.textContent = 'Indisponível';
    if (messageElement) {
      messageElement.textContent = 'Não foi possível carregar o catálogo do Corsarr.';
      messageElement.classList.add('error');
    }
  }
}

function renderLegalNotices(): void {
  if (!legalNoticesElement) return;
  legalNoticesElement.replaceChildren(
    ...legalNotices.map((notice) => {
      const card = document.createElement('article');
      card.className = 'legal-card';

      const heading = document.createElement('div');
      heading.className = 'legal-heading';
      const title = document.createElement('h2');
      title.textContent = notice.name;
      const kind = document.createElement('span');
      kind.textContent = notice.componentType === 'runtime' ? 'Infraestrutura' : 'Aplicativo';
      heading.append(title, kind);

      const purpose = document.createElement('p');
      purpose.textContent = notice.purpose;
      const license = document.createElement('strong');
      license.textContent = notice.license;
      const copyright = document.createElement('p');
      copyright.textContent = notice.copyrightNotice;

      card.append(heading, purpose, license, copyright);
      if (notice.imageMaintainer) {
        const image = document.createElement('p');
        image.className = 'legal-image';
        const installedImage = managedStatuses.get(notice.id)?.image;
        image.textContent = `Imagem mantida por ${notice.imageMaintainer} · ${installedImage ? 'Instalada' : 'Aprovada'}: ${installedImage ?? notice.approvedImage}`;
        card.append(image);
      }

      const affiliation = document.createElement('p');
      affiliation.className = 'legal-affiliation';
      affiliation.textContent = notice.affiliationStatement;
      card.append(affiliation);

      const links = document.createElement('div');
      links.className = 'legal-links';
      for (const link of notice.links) {
        const button = document.createElement('button');
        button.type = 'button';
        button.textContent = link.label;
        button.addEventListener('click', async () => {
          button.disabled = true;
          try {
            await OpenLegalLink(notice.id, link.kind);
          } finally {
            button.disabled = false;
          }
        });
        links.append(button);
      }
      card.append(links);
      return card;
    }),
  );
}

async function loadLegalNotices(): Promise<void> {
  try {
    legalNotices = await ListLegalNotices();
    renderLegalNotices();
  } catch {
    if (legalNoticesElement) {
      legalNoticesElement.textContent = 'Não foi possível carregar os créditos e licenças.';
    }
  }
}

async function loadApplicationStatuses(): Promise<void> {
  try {
    const statuses = await GetApplicationStatuses();
    managedStatuses = new Map(statuses.map((status) => [status.applicationId, status]));
    renderApplications();
    renderLegalNotices();
  } catch {
    managedStatuses = new Map();
  }
}

async function loadApplicationDataStatuses(): Promise<void> {
  try {
    const statuses = await GetApplicationDataStatuses();
    dataStatuses = new Map(statuses.map((status) => [status.applicationId, status]));
    renderApplications();
  } catch {
    dataStatuses = new Map();
  }
}

async function loadQBittorrentAccess(): Promise<void> {
  try {
    qbittorrentAccess = await GetQBittorrentAccessStatus();
    renderApplications();
  } catch {
    qbittorrentAccess = undefined;
  }
}

async function loadJellyfinAccess(): Promise<void> {
  try {
    jellyfinAccess = await GetJellyfinAccessStatus();
    renderApplications();
  } catch {
    jellyfinAccess = undefined;
  }
}

function applySetupStatus(status: application.SetupStatus): void {
  setupStatus = status;
  selectedApplicationIDs = new Set(status.applications);

  if (status.storagePath) {
    if (storageTitleElement) storageTitleElement.textContent = 'Pasta salva';
    if (storageDescriptionElement) {
      storageDescriptionElement.textContent =
        'O Corsarr verificará novamente esta pasta antes de preparar os aplicativos.';
    }
    if (storagePathElement) storagePathElement.textContent = status.storagePath;
    if (storageFactsElement) storageFactsElement.textContent = '';
    if (storageBadgeElement) {
      storageBadgeElement.textContent = 'Salva';
      storageBadgeElement.className = 'runtime-badge ready';
    }
    if (chooseStorageButton) chooseStorageButton.textContent = 'Trocar pasta';
  }

  if (installationSummaryElement) {
    if (status.canInstall) {
      const applicationLabel =
        status.applications.length === 1
          ? '1 aplicativo selecionado'
          : `${status.applications.length} aplicativos selecionados`;
      installationSummaryElement.textContent = `${applicationLabel}. As dependências necessárias são incluídas automaticamente.`;
    } else if (!status.storagePath && status.applications.length === 0) {
      installationSummaryElement.textContent = 'Escolha uma pasta e ao menos um aplicativo.';
    } else if (!status.storagePath) {
      installationSummaryElement.textContent = 'Agora escolha a pasta onde a mídia será guardada.';
    } else {
      installationSummaryElement.textContent = 'Agora escolha ao menos um aplicativo.';
    }
  }
  if (prepareStorageButton) prepareStorageButton.disabled = !status.canPrepare;
  if (acceptTermsCheckbox) acceptTermsCheckbox.checked = status.termsAccepted;
  updateInstallAuthority();
}

function updateInstallAuthority(): void {
  if (!installApplicationsButton) return;
  installApplicationsButton.disabled = !setupStatus?.canPrepare || !acceptTermsCheckbox?.checked;
}

async function loadSetup(): Promise<void> {
  try {
    applySetupStatus(await GetSetupStatus());
  } catch {
    if (installationSummaryElement) {
      installationSummaryElement.textContent =
        'Não foi possível recuperar a preparação salva neste computador.';
    }
    if (prepareStorageButton) prepareStorageButton.disabled = true;
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
    currentRuntimeState = environment.runtime.state;
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
    if (prepareRuntimeButton) {
      prepareRuntimeButton.hidden = environment.runtime.state === 'ready';
      prepareRuntimeButton.textContent =
        environment.runtime.state === 'stopped' ? 'Iniciar ambiente' : 'Preparar computador';
    }
  } catch {
    currentRuntimeState = 'error';
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

async function prepareRuntime(): Promise<void> {
  if (!prepareRuntimeButton) return;
  if (!acceptTermsCheckbox?.checked) {
    if (installationResultElement) {
      installationResultElement.textContent =
        'Leia e marque a autorização abaixo antes de preparar este computador.';
      installationResultElement.classList.add('error');
      acceptTermsCheckbox?.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
    return;
  }
  if (
    currentRuntimeState === 'unavailable' &&
    !window.confirm(
      'O Corsarr baixará o Docker Desktop 4.86.0 diretamente da Docker, verificará o checksum e a assinatura oficial e solicitará a autorização do macOS para instalar. Continuar?',
    )
  ) {
    return;
  }

  prepareRuntimeButton.disabled = true;
  prepareRuntimeButton.textContent =
    currentRuntimeState === 'unavailable' ? 'Instalando…' : 'Iniciando…';
  if (environmentDescriptionElement) {
    environmentDescriptionElement.textContent =
      currentRuntimeState === 'unavailable'
        ? 'Baixando e verificando os componentes oficiais. O macOS poderá solicitar sua senha.'
        : 'Iniciando os componentes necessários. Isso pode levar alguns instantes.';
  }
  try {
    applySetupStatus(await AcceptCurrentTerms());
    const result = await PrepareRuntime();
    if (!result.ready) throw new Error('runtime not ready');
    await loadEnvironment();
    if (messageElement) {
      messageElement.textContent = result.installed
        ? 'Este computador foi preparado e está pronto para instalar os aplicativos.'
        : 'O ambiente foi iniciado e está pronto.';
      messageElement.classList.remove('error');
    }
  } catch {
    if (environmentDescriptionElement) {
      environmentDescriptionElement.textContent =
        'A preparação não terminou. Nada foi instalado sem assinatura válida; tente novamente ou consulte os detalhes técnicos.';
    }
    if (messageElement) {
      messageElement.textContent = 'Não foi possível concluir a preparação deste computador.';
      messageElement.classList.add('error');
    }
  } finally {
    prepareRuntimeButton.disabled = false;
    if (currentRuntimeState !== 'ready') {
      prepareRuntimeButton.textContent =
        currentRuntimeState === 'stopped' ? 'Iniciar ambiente' : 'Preparar computador';
    }
  }
}

prepareRuntimeButton?.addEventListener('click', () => void prepareRuntime());

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
      applySetupStatus(await GetSetupStatus());
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

async function prepareStorage(): Promise<void> {
  if (!prepareStorageButton || !setupStatus?.canPrepare) return;

  prepareStorageButton.disabled = true;
  prepareStorageButton.textContent = 'Preparando…';
  if (installationResultElement) {
    installationResultElement.textContent = 'Criando somente as pastas revisadas acima…';
    installationResultElement.classList.remove('error');
  }

  try {
    const layout = await PrepareStorageLayout();
    if (installationResultElement) {
      installationResultElement.textContent = `Estrutura pronta em ${layout.rootPath}. Nenhum aplicativo foi instalado ainda.`;
    }
    prepareStorageButton.textContent = 'Estrutura pronta';
  } catch {
    if (installationResultElement) {
      installationResultElement.textContent =
        'Não foi possível criar as pastas. Sua seleção foi preservada para uma nova tentativa.';
      installationResultElement.classList.add('error');
    }
    prepareStorageButton.textContent = 'Tentar novamente';
  } finally {
    prepareStorageButton.disabled = !setupStatus?.canPrepare;
  }
}

prepareStorageButton?.addEventListener('click', () => void prepareStorage());
acceptTermsCheckbox?.addEventListener('change', updateInstallAuthority);

async function installApplications(): Promise<void> {
  if (!installApplicationsButton || !setupStatus?.canPrepare || !acceptTermsCheckbox?.checked) {
    return;
  }

  installApplicationsButton.disabled = true;
  installApplicationsButton.textContent = 'Instalando…';
  if (installationResultElement) {
    installationResultElement.textContent =
      'Preparando o ambiente e baixando os aplicativos. Isso pode levar alguns minutos.';
    installationResultElement.classList.remove('error');
  }

  try {
    applySetupStatus(await AcceptCurrentTerms());
    const result = await InstallSelectedApplications();
    if (result.complete) {
      if (installationResultElement) {
        installationResultElement.textContent = `${result.items.length} aplicativos instalados e iniciados.`;
      }
      installApplicationsButton.textContent = 'Aplicativos instalados';
      await Promise.all([
        loadApplicationStatuses(),
        loadApplicationDataStatuses(),
        loadJellyfinAccess(),
        loadQBittorrentAccess(),
      ]);
    } else {
      const failed = result.items.find((item) => item.error);
      if (installationResultElement) {
        installationResultElement.textContent = failed
          ? `A instalação de ${failed.applicationId} não terminou. Tente novamente ou consulte os detalhes técnicos.`
          : 'A instalação não terminou. Tente novamente.';
        installationResultElement.classList.add('error');
      }
      installApplicationsButton.textContent = 'Tentar novamente';
    }
  } catch {
    if (installationResultElement) {
      installationResultElement.textContent =
        'Não foi possível iniciar a instalação. Sua pasta e seleção continuam preservadas.';
      installationResultElement.classList.add('error');
    }
    installApplicationsButton.textContent = 'Tentar novamente';
  } finally {
    updateInstallAuthority();
  }
}

installApplicationsButton?.addEventListener('click', () => void installApplications());

async function loadInitialState(): Promise<void> {
  await Promise.all([loadEnvironment(), loadSetup()]);
  await Promise.all([loadApplications(), loadLegalNotices()]);
  await Promise.all([
    loadApplicationStatuses(),
    loadApplicationDataStatuses(),
    loadJellyfinAccess(),
    loadQBittorrentAccess(),
  ]);
}

void loadInitialState();
