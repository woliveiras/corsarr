import './style.css';
import {
  AcceptCurrentTerms,
  AdvanceOnboarding,
  ArchiveApplicationData,
  ChooseStorageLocation,
  CopyJellyfinNetworkURL,
  CopyJellyfinPassword,
  CopyQBittorrentPassword,
  ExportDiagnostics,
  GetApplicationDataStatuses,
  GetApplicationStatuses,
  GetEnvironmentStatus,
  GetJellyfinAccessStatus,
  GetJellyfinNetworkStatus,
  GetQBittorrentAccessStatus,
  GetSetupStatus,
  InstallSelectedApplications,
  ListApplications,
  ListLegalNotices,
  OpenApplication,
  OpenLegalLink,
  OpenStartAtLoginSettings,
  PrepareRuntime,
  PrepareStorageLayout,
  RemoveApplication,
  RestartApplication,
  SaveApplicationSelection,
  SelectRecommendedApplications,
  SetJellyfinLAN,
  SetStartAtLogin,
  StartApplication,
  StopApplication,
  UpdateApplication,
} from '../wailsjs/go/main/App';
import type { application, legal, main, storage } from '../wailsjs/go/models';
import { EventsOn } from '../wailsjs/runtime/runtime';

type Application = application.ApplicationSummary;
type ManagedStatus = application.ManagedApplicationStatus;
type DataStatus = storage.ApplicationDataStatus;
type LegalNotice = legal.Notice;
type InstallationProgress = {
  applicationId: string;
  stage: 'installing' | 'provisioning' | 'ready' | 'failed';
  position: number;
  total: number;
};

const root = document.querySelector<HTMLDivElement>('#app');

if (!root) {
  throw new Error('Elemento principal do Corsarr não encontrado.');
}

root.innerHTML = [
  '<section id="onboarding" class="onboarding" hidden aria-label="Configuração inicial do Corsarr">',
  '  <div class="onboarding-frame">',
  '    <header class="onboarding-header">',
  '      <div class="onboarding-brand"><span class="brand-mark" aria-hidden="true">C</span><span><strong>Corsarr</strong><small>Configuração inicial</small></span></div>',
  '      <div id="onboarding-progress" class="onboarding-progress" hidden><span class="active"></span><span></span><span></span><span></span><small id="onboarding-progress-label">Etapa 1 de 4</small></div>',
  '    </header>',
  '    <article id="onboarding-splash" class="onboarding-step onboarding-splash">',
  '      <div class="onboarding-splash-copy"><p class="eyebrow">BEM-VINDO AO CORSARR</p><h1>Seu servidor de mídia,<br><em>sem complicação.</em></h1><p>Vamos preparar este computador juntos. Você entenderá cada escolha antes que o Corsarr instale qualquer componente.</p><button id="onboarding-start" class="onboarding-primary" type="button">Começar configuração</button></div>',
  '      <div class="radar onboarding-radar" aria-hidden="true"><span></span><span></span><i></i><b>C</b></div>',
  '    </article>',
  '    <article id="onboarding-permissions" class="onboarding-step" hidden>',
  '      <div class="onboarding-step-copy"><p class="eyebrow">ETAPA 1 DE 4 · AUTORIZAÇÃO</p><h1>Você mantém o controle.</h1><p>O Corsarr usará o Docker Desktop para executar os aplicativos em serviços isolados. Ele só instalará componentes, baixará imagens e criará serviços depois da sua autorização.</p><div class="onboarding-explanation"><strong>O que será autorizado</strong><ul><li>Usar ou instalar o Docker Desktop neste Mac.</li><li>Baixar somente imagens aprovadas e identificadas por versão.</li><li>Criar pastas e serviços apenas dentro da configuração do Corsarr.</li></ul></div><label class="onboarding-check"><input id="onboarding-terms" type="checkbox"><span>Li e autorizo o uso do Docker Desktop e dos aplicativos selecionados. Entendo que cada componente mantém sua própria licença.</span></label><button id="onboarding-open-docker-terms" class="onboarding-link" type="button">Abrir termos oficiais do Docker Desktop</button><label id="onboarding-login-setting" class="onboarding-check"><input id="onboarding-start-login" type="checkbox"><span>Iniciar meus serviços automaticamente quando eu entrar neste Mac.</span></label><p id="onboarding-permissions-message" class="onboarding-message"></p></div>',
  '      <footer class="onboarding-actions"><button class="onboarding-back" type="button" data-onboarding-step="splash">Voltar</button><button id="onboarding-permissions-next" class="onboarding-primary" type="button" disabled>Autorizar e continuar</button></footer>',
  '    </article>',
  '    <article id="onboarding-environment" class="onboarding-step" hidden>',
  '      <div class="onboarding-step-copy"><p class="eyebrow">ETAPA 2 DE 4 · AMBIENTE</p><h1>Preparando este computador.</h1><p>O Docker mantém cada aplicativo separado e permite que o Corsarr cuide de instalação, atualização e reinício por você.</p><div class="onboarding-diagnostic"><span class="environment-icon" aria-hidden="true">◎</span><div><small>DIAGNÓSTICO</small><strong id="onboarding-environment-title">Verificando…</strong><p id="onboarding-environment-description">Aguarde enquanto conferimos este Mac.</p><details><summary>Detalhes técnicos</summary><code id="onboarding-environment-technical"></code></details></div><span id="onboarding-environment-badge" class="runtime-badge checking">Verificando</span></div><p id="onboarding-environment-message" class="onboarding-message"></p></div>',
  '      <footer class="onboarding-actions"><button class="onboarding-back" type="button" data-onboarding-step="permissions">Voltar</button><div><button id="onboarding-prepare-runtime" class="secondary-button" type="button" hidden>Preparar computador</button><button id="onboarding-environment-next" class="onboarding-primary" type="button" disabled>Próximo</button></div></footer>',
  '    </article>',
  '    <article id="onboarding-storage" class="onboarding-step" hidden>',
  '      <div class="onboarding-step-copy"><p class="eyebrow">ETAPA 3 DE 4 · ARMAZENAMENTO</p><h1>Onde sua mídia ficará?</h1><p>Dentro da pasta escolhida, o Corsarr criará uma estrutura clara para configurações, downloads e bibliotecas. Você poderá encontrá-la depois sem conhecer containers.</p><div class="onboarding-path-preview"><span class="storage-icon" aria-hidden="true">▱</span><div><strong id="onboarding-storage-title">Nenhuma pasta selecionada</strong><p id="onboarding-storage-description">Escolha uma pasta com pelo menos 10 GB disponíveis.</p><code id="onboarding-storage-path"></code><small id="onboarding-storage-facts"></small></div><span id="onboarding-storage-badge" class="runtime-badge checking">Pendente</span></div><div class="folder-tree" aria-label="Estrutura que será criada"><code>Corsarr/<br>├── config/ <span>configurações dos aplicativos</span><br>└── media/<br>&nbsp;&nbsp;&nbsp;├── downloads/ <span>arquivos baixados</span><br>&nbsp;&nbsp;&nbsp;└── library/ <span>filmes, séries, músicas e livros</span></code></div><p id="onboarding-storage-message" class="onboarding-message"></p></div>',
  '      <footer class="onboarding-actions"><button class="onboarding-back" type="button" data-onboarding-step="environment">Voltar</button><div><button id="onboarding-choose-storage" class="secondary-button" type="button">Escolher pasta</button><button id="onboarding-storage-next" class="onboarding-primary" type="button" disabled>Próximo</button></div></footer>',
  '    </article>',
  '    <article id="onboarding-applications" class="onboarding-step" hidden>',
  '      <div class="onboarding-step-copy onboarding-applications-copy"><p class="eyebrow">ETAPA 4 DE 4 · APLICATIVOS</p><h1>Escolha o que deseja instalar.</h1><p>O Corsarr incluirá automaticamente as dependências necessárias. Cada aplicativo será executado em seu próprio container e continuará disponível pela interface original.</p><div class="onboarding-catalog-heading"><span id="onboarding-catalog-count">Carregando…</span><button id="onboarding-recommended" class="secondary-button" type="button">Usar configuração recomendada</button></div><div id="onboarding-application-list" class="onboarding-application-list"></div><label id="onboarding-jellyfin-lan-setting" class="onboarding-check" hidden><input id="onboarding-jellyfin-lan" type="checkbox"><span>Permitir assistir no Jellyfin por TVs e aparelhos desta rede local.</span></label><p id="onboarding-installation-result" class="onboarding-message" aria-live="polite"></p><details id="onboarding-operation-details" class="operation-details" hidden><summary>Detalhes técnicos</summary><code id="onboarding-operation-technical"></code></details></div>',
  '      <footer class="onboarding-actions"><button class="onboarding-back" type="button" data-onboarding-step="storage">Voltar</button><button id="onboarding-install" class="onboarding-primary" type="button" disabled>Instalar aplicativos</button></footer>',
  '    </article>',
  '  </div>',
  '</section>',
  '<div id="dashboard-shell" class="shell" hidden>',
  '  <aside class="sidebar">',
  '    <a class="brand" href="#" aria-label="Corsarr, início">',
  '      <span class="brand-mark" aria-hidden="true">C</span>',
  '      <span><strong>Corsarr</strong><small>Desktop</small></span>',
  '    </a>',
  '    <nav aria-label="Navegação principal">',
  '      <button id="show-home" class="nav-item active" type="button"><span aria-hidden="true">⌂</span>Início</button>',
  '      <button id="show-licenses" class="nav-item" type="button"><span aria-hidden="true">§</span>Licenças</button>',
  '      <button id="export-diagnostics" class="nav-item" type="button"><span aria-hidden="true">⇩</span>Exportar diagnóstico</button>',
  '    </nav>',
  '    <div class="sidebar-note">',
  '      <span class="status-dot"></span>',
  '      <p><strong>Tudo em um só lugar</strong>Instale e cuide dos seus aplicativos sem precisar usar o terminal.</p>',
  '    </div>',
  '  </aside>',
  '  <main class="content">',
  '    <div id="home-view">',
  '    <header class="topbar">',
  '      <div><p class="eyebrow">SEU SERVIDOR DE MÍDIA</p><h1>Olá.</h1></div>',
  '      <div class="machine"><span class="machine-icon" aria-hidden="true">⌘</span><span><small id="platform-name">Verificando computador</small><strong id="architecture-name">Aguarde…</strong></span></div>',
  '    </header>',
  '    <section class="hero" aria-labelledby="hero-title">',
  '      <div>',
  '        <span class="phase">CONTROLE LOCAL</span>',
  '        <h2 id="hero-title">Tudo no lugar certo,<br><em>sem complicação.</em></h2>',
  '        <p>Instale, atualize e cuide dos seus aplicativos de mídia por uma única interface.</p>',
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
  '      <div><p class="eyebrow">APLICATIVOS</p><h2>Seus aplicativos</h2><p class="section-help">Não sabe por onde começar? O Corsarr pode escolher uma configuração completa para filmes e séries.</p></div>',
  '      <div class="catalog-actions"><p id="catalog-count" class="catalog-count">Carregando…</p><button id="select-recommended" class="secondary-button recommended-button" type="button">Usar configuração recomendada</button></div>',
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
  '        <div id="start-at-login-setting" class="start-at-login-setting" hidden><label><input id="start-at-login" type="checkbox"> <span>Iniciar meus serviços automaticamente quando eu entrar no Mac.</span></label><button id="open-login-settings" class="legal-link-button" type="button" hidden>Abrir Ajustes do Sistema</button></div>',
  '        <div id="jellyfin-lan-setting" class="start-at-login-setting" hidden><label><input id="jellyfin-lan" type="checkbox"> <span>Permitir assistir no Jellyfin por TVs e aparelhos desta rede local. Os painéis administrativos continuam privados neste computador.</span></label></div>',
  '        <p id="installation-result" class="installation-result"></p>',
  '        <details id="operation-details" class="operation-details" hidden><summary>Detalhes técnicos</summary><code id="operation-technical"></code></details>',
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
const onboardingElement = document.querySelector<HTMLElement>('#onboarding');
const dashboardShell = document.querySelector<HTMLElement>('#dashboard-shell');
const onboardingProgress = document.querySelector<HTMLElement>('#onboarding-progress');
const onboardingProgressLabel = document.querySelector<HTMLElement>('#onboarding-progress-label');
const onboardingStartButton = document.querySelector<HTMLButtonElement>('#onboarding-start');
const onboardingTermsCheckbox = document.querySelector<HTMLInputElement>('#onboarding-terms');
const onboardingStartLoginCheckbox =
  document.querySelector<HTMLInputElement>('#onboarding-start-login');
const onboardingPermissionsNext = document.querySelector<HTMLButtonElement>(
  '#onboarding-permissions-next',
);
const onboardingOpenDockerTerms = document.querySelector<HTMLButtonElement>(
  '#onboarding-open-docker-terms',
);
const onboardingPermissionsMessage = document.querySelector<HTMLElement>(
  '#onboarding-permissions-message',
);
const onboardingEnvironmentTitle = document.querySelector<HTMLElement>(
  '#onboarding-environment-title',
);
const onboardingEnvironmentDescription = document.querySelector<HTMLElement>(
  '#onboarding-environment-description',
);
const onboardingEnvironmentTechnical = document.querySelector<HTMLElement>(
  '#onboarding-environment-technical',
);
const onboardingEnvironmentBadge = document.querySelector<HTMLElement>(
  '#onboarding-environment-badge',
);
const onboardingPrepareRuntime = document.querySelector<HTMLButtonElement>(
  '#onboarding-prepare-runtime',
);
const onboardingEnvironmentNext = document.querySelector<HTMLButtonElement>(
  '#onboarding-environment-next',
);
const onboardingEnvironmentMessage = document.querySelector<HTMLElement>(
  '#onboarding-environment-message',
);
const onboardingStorageTitle = document.querySelector<HTMLElement>('#onboarding-storage-title');
const onboardingStorageDescription = document.querySelector<HTMLElement>(
  '#onboarding-storage-description',
);
const onboardingStoragePath = document.querySelector<HTMLElement>('#onboarding-storage-path');
const onboardingStorageFacts = document.querySelector<HTMLElement>('#onboarding-storage-facts');
const onboardingStorageBadge = document.querySelector<HTMLElement>('#onboarding-storage-badge');
const onboardingChooseStorage = document.querySelector<HTMLButtonElement>(
  '#onboarding-choose-storage',
);
const onboardingStorageNext = document.querySelector<HTMLButtonElement>('#onboarding-storage-next');
const onboardingStorageMessage = document.querySelector<HTMLElement>('#onboarding-storage-message');
const onboardingCatalogCount = document.querySelector<HTMLElement>('#onboarding-catalog-count');
const onboardingRecommended = document.querySelector<HTMLButtonElement>('#onboarding-recommended');
const onboardingApplicationList = document.querySelector<HTMLElement>(
  '#onboarding-application-list',
);
const onboardingJellyfinLANSetting = document.querySelector<HTMLElement>(
  '#onboarding-jellyfin-lan-setting',
);
const onboardingJellyfinLANCheckbox = document.querySelector<HTMLInputElement>(
  '#onboarding-jellyfin-lan',
);
const onboardingInstallButton = document.querySelector<HTMLButtonElement>('#onboarding-install');
const onboardingInstallationResult = document.querySelector<HTMLElement>(
  '#onboarding-installation-result',
);
const onboardingOperationDetails = document.querySelector<HTMLDetailsElement>(
  '#onboarding-operation-details',
);
const onboardingOperationTechnical = document.querySelector<HTMLElement>(
  '#onboarding-operation-technical',
);
const countElement = document.querySelector<HTMLElement>('#catalog-count');
const selectRecommendedButton = document.querySelector<HTMLButtonElement>('#select-recommended');
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
const operationDetailsElement = document.querySelector<HTMLDetailsElement>('#operation-details');
const operationTechnicalElement = document.querySelector<HTMLElement>('#operation-technical');
const prepareStorageButton = document.querySelector<HTMLButtonElement>('#prepare-storage');
const installApplicationsButton =
  document.querySelector<HTMLButtonElement>('#install-applications');
const acceptTermsCheckbox = document.querySelector<HTMLInputElement>('#accept-terms');
const startAtLoginSetting = document.querySelector<HTMLElement>('#start-at-login-setting');
const startAtLoginCheckbox = document.querySelector<HTMLInputElement>('#start-at-login');
const openLoginSettingsButton = document.querySelector<HTMLButtonElement>('#open-login-settings');
const jellyfinLANSetting = document.querySelector<HTMLElement>('#jellyfin-lan-setting');
const jellyfinLANCheckbox = document.querySelector<HTMLInputElement>('#jellyfin-lan');
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
let jellyfinNetwork: main.JellyfinNetworkStatus | undefined;
let legalNotices: LegalNotice[] = [];
let currentRuntimeState = 'checking';
let currentHostReady = true;

type OnboardingStep = 'splash' | 'permissions' | 'environment' | 'storage' | 'applications';

const onboardingStepNumbers: Partial<Record<OnboardingStep, number>> = {
  permissions: 1,
  environment: 2,
  storage: 3,
  applications: 4,
};

const onboardingStepOrder: Record<OnboardingStep, number> = {
  splash: 0,
  permissions: 1,
  environment: 2,
  storage: 3,
  applications: 4,
};

function persistedOnboardingStep(): OnboardingStep {
  const step = setupStatus?.onboardingStep;
  if (!step || step === 'welcome' || step === 'complete') return 'splash';
  return step as OnboardingStep;
}

function onboardingHasAdvancedPast(step: OnboardingStep): boolean {
  return onboardingStepOrder[persistedOnboardingStep()] > onboardingStepOrder[step];
}

function showOnboardingStep(step: OnboardingStep): void {
  for (const element of document.querySelectorAll<HTMLElement>('.onboarding-step')) {
    const active = element.id === `onboarding-${step}`;
    element.hidden = !active;
    element.inert = !active;
    element.setAttribute('aria-hidden', String(!active));
  }
  const stepNumber = onboardingStepNumbers[step];
  if (onboardingProgress) onboardingProgress.hidden = stepNumber === undefined;
  if (onboardingProgressLabel && stepNumber !== undefined) {
    onboardingProgressLabel.textContent = `Etapa ${stepNumber} de 4`;
  }
  if (onboardingProgress && stepNumber !== undefined) {
    const indicators = onboardingProgress.querySelectorAll<HTMLSpanElement>('span');
    for (const [index, indicator] of [...indicators].entries()) {
      indicator.classList.toggle('active', index < stepNumber);
    }
  }
  document.querySelector('.onboarding-frame')?.scrollTo({ top: 0, behavior: 'smooth' });
}

function syncOnboarding(status: application.SetupStatus): void {
  const completed = status.onboardingCompleted;
  if (onboardingElement) onboardingElement.hidden = completed;
  if (dashboardShell) dashboardShell.hidden = !completed;
  if (completed) return;

  const step = (status.onboardingStep || 'welcome') as string;
  const visibleStep: OnboardingStep =
    step === 'welcome' || step === 'complete' ? 'splash' : (step as OnboardingStep);
  showOnboardingStep(visibleStep);
  if (onboardingTermsCheckbox) onboardingTermsCheckbox.checked = status.termsAccepted;
  if (onboardingStartLoginCheckbox) {
    onboardingStartLoginCheckbox.checked = status.startAtLogin;
    onboardingStartLoginCheckbox.disabled = !status.startAtLoginSupported;
  }
  if (onboardingPermissionsNext) {
    onboardingPermissionsNext.disabled = !onboardingTermsCheckbox?.checked;
  }
  if (onboardingStorageNext) onboardingStorageNext.disabled = !status.storagePath;
  if (onboardingInstallButton) onboardingInstallButton.disabled = !status.canInstall;
  if (onboardingJellyfinLANSetting) {
    onboardingJellyfinLANSetting.hidden = !status.applications.includes('jellyfin');
  }
  if (onboardingJellyfinLANCheckbox) {
    onboardingJellyfinLANCheckbox.checked = status.jellyfinLanEnabled;
  }
}

function renderOnboardingIssue(issue?: application.OperationIssue): void {
  if (!onboardingOperationDetails || !onboardingOperationTechnical) return;
  onboardingOperationDetails.hidden = !issue;
  onboardingOperationDetails.open = false;
  onboardingOperationTechnical.textContent = issue
    ? `${issue.summary}\n${issue.nextAction}\nCódigo: ${issue.code}`
    : '';
}

function renderOperationIssue(issue?: application.OperationIssue): void {
  if (!operationDetailsElement || !operationTechnicalElement) return;
  operationDetailsElement.hidden = !issue;
  operationDetailsElement.open = false;
  operationTechnicalElement.textContent = issue
    ? `${issue.summary}\n${issue.nextAction}\nCódigo: ${issue.code}`
    : '';
}

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

selectRecommendedButton?.addEventListener('click', async () => {
  if (selectionSaving || !selectRecommendedButton) return;
  selectionSaving = true;
  selectRecommendedButton.disabled = true;
  renderApplications();

  try {
    const status = await SelectRecommendedApplications();
    applySetupStatus(status);
    if (messageElement) {
      messageElement.textContent =
        'Configuração recomendada selecionada. Você ainda pode adicionar ou remover aplicativos.';
      messageElement.classList.remove('error');
    }
  } catch {
    if (messageElement) {
      messageElement.textContent = 'Não foi possível selecionar a configuração recomendada.';
      messageElement.classList.add('error');
    }
  } finally {
    selectionSaving = false;
    selectRecommendedButton.disabled = false;
    renderApplications();
  }
});

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
  const managedStatus = managedStatuses.get(application.id);
  const installed = managedStatus?.state === 'running' || managedStatus?.state === 'stopped';
  const uncertainSelection =
    managedStatus?.state === 'attention' && selectedApplicationIDs.has(application.id);
  const selected = selectedApplicationIDs.has(application.id) || installed;
  if (selected) card.classList.add('selected');

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
  const stateLabels: Record<string, string> = {
    not_installed: 'Não instalado',
    running: 'Em execução',
    stopped: 'Parado',
    attention: 'Atenção',
  };
  metadata.textContent = `${application.optional ? 'Opcional' : 'Aplicativo principal'} · ${stateLabels[managedStatus?.state ?? 'not_installed']}${managedStatus?.updateAvailable ? ' · Atualização disponível' : ''}`;
  if ((managedStatus?.removalBlockedBy?.length ?? 0) > 0) {
    const blockerNames = managedStatus?.removalBlockedBy?.map(
      (id) => availableApplications.find((candidate) => candidate.id === id)?.name ?? id,
    );
    metadata.textContent += ` · Remova primeiro: ${blockerNames?.join(', ')}`;
  }

  information.append(title, description, metadata);

  if (managedStatus?.issue) {
    const details = document.createElement('details');
    details.className = 'application-status-details';
    const summary = document.createElement('summary');
    summary.textContent = 'Ver detalhes';
    const diagnostic = document.createElement('code');
    diagnostic.textContent = `${managedStatus.issue.summary}\n${managedStatus.issue.nextAction}\nCódigo: ${managedStatus.issue.code}`;
    details.append(summary, diagnostic);
    information.append(details);
  }

  if (
    application.id === 'jellyfin' &&
    managedStatus?.state === 'running' &&
    jellyfinNetwork?.enabled &&
    jellyfinNetwork.urls.length > 0
  ) {
    const networkAddress = document.createElement('p');
    networkAddress.className = 'network-address';
    networkAddress.textContent = `TV e celular: ${jellyfinNetwork.urls[0]}`;
    networkAddress.title = jellyfinNetwork.urls.join('\n');
    information.append(networkAddress);
  }

  const actions = document.createElement('div');
  actions.className = 'application-actions';

  const selectButton = document.createElement('button');
  selectButton.className = 'select-button';
  selectButton.type = 'button';
  selectButton.textContent = installed ? 'Instalado' : selected ? 'Selecionado' : 'Selecionar';
  selectButton.setAttribute('aria-pressed', String(selected));
  selectButton.setAttribute('aria-label', `Selecionar ${application.name} para instalação`);
  selectButton.disabled = selectionSaving || installed || uncertainSelection;
  if (installed) {
    selectButton.title = 'Remova o aplicativo antes de retirá-lo da seleção.';
    selectButton.setAttribute(
      'aria-label',
      `${application.name} está instalado. Remova o aplicativo antes de retirá-lo da seleção.`,
    );
  } else if (uncertainSelection) {
    selectButton.title = 'Verifique o ambiente antes de retirar este aplicativo da seleção.';
    selectButton.setAttribute(
      'aria-label',
      `${application.name} continua selecionado enquanto o ambiente não pode ser verificado.`,
    );
  }
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
      applicationRemovalButton(application, managedStatus),
    );
  } else if (managedStatus?.state === 'stopped') {
    actions.append(
      lifecycleButton('Iniciar', () => StartApplication(application.id)),
      applicationRemovalButton(application, managedStatus),
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
  if (
    application.id === 'jellyfin' &&
    managedStatus?.state === 'running' &&
    jellyfinNetwork?.enabled &&
    jellyfinNetwork.urls.length > 0
  ) {
    actions.append(jellyfinNetworkButton(jellyfinNetwork.urls[0]));
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
      renderOperationIssue(result.issue);
      if (messageElement) {
        if (result.updated && !result.requiresAttention) {
          messageElement.textContent = `${target.name} foi atualizado e verificado. O backup das configurações foi preservado.`;
          messageElement.classList.remove('error');
        } else if (result.rolledBack) {
          messageElement.textContent = `A nova versão de ${target.name} não passou na verificação. A imagem anterior foi restaurada e o backup foi preservado.`;
          messageElement.classList.add('error');
        } else if (result.requiresAttention) {
          messageElement.textContent = `${target.name} requer atenção após a tentativa de atualização. Consulte os detalhes técnicos.`;
          messageElement.classList.add('error');
        } else {
          messageElement.textContent = `${target.name} já usa a versão aprovada pelo Corsarr.`;
          messageElement.classList.remove('error');
        }
      }
      await loadApplicationStatuses();
    } catch {
      renderOperationIssue();
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

function applicationRemovalButton(target: Application, status: ManagedStatus): HTMLButtonElement {
  const button = lifecycleButton('Remover', () => removeApplication(target));
  const blockers = status.removalBlockedBy ?? [];
  if (blockers.length === 0) return button;

  const blockerNames = blockers.map(
    (id) => availableApplications.find((application) => application.id === id)?.name ?? id,
  );
  button.disabled = true;
  button.title = `Remova primeiro: ${blockerNames.join(', ')}.`;
  button.setAttribute('aria-label', `Não é possível remover ${target.name}. ${button.title}`);
  return button;
}

function dataRemovalButton(target: Application): HTMLButtonElement {
  const button = document.createElement('button');
  button.className = 'data-removal-button';
  button.type = 'button';
  button.textContent = 'Remover dados';
  button.addEventListener('click', async () => {
    const dataStatus = dataStatuses.get(target.id);
    const approximateSize = formatApproximateBytes(dataStatus?.sizeBytes ?? 0);
    const confirmed = window.confirm(
      `Remover aproximadamente ${approximateSize} de configurações de ${target.name}? A biblioteca e os downloads não serão alterados. A configuração será movida para a lixeira do Corsarr e poderá ser recuperada manualmente.`,
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

function formatApproximateBytes(bytes: number): string {
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = Math.max(0, bytes);
  let unitIndex = 0;
  while (value >= 1000 && unitIndex < units.length - 1) {
    value /= 1000;
    unitIndex += 1;
  }
  const maximumFractionDigits = unitIndex === 0 ? 0 : 1;
  return `${new Intl.NumberFormat('pt-BR', { maximumFractionDigits }).format(value)} ${units[unitIndex]}`;
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

function jellyfinNetworkButton(url: string): HTMLButtonElement {
  const button = document.createElement('button');
  button.className = 'network-button';
  button.type = 'button';
  button.textContent = 'Copiar endereço';
  button.addEventListener('click', async () => {
    button.disabled = true;
    try {
      await CopyJellyfinNetworkURL(url);
      if (messageElement) {
        messageElement.textContent =
          'Endereço copiado. Abra-o em um aparelho conectado à mesma rede local.';
        messageElement.classList.remove('error');
      }
    } catch {
      if (messageElement) {
        messageElement.textContent =
          'Não foi possível copiar o endereço. Verifique se o acesso pela rede continua ativado.';
        messageElement.classList.add('error');
      }
    } finally {
      button.disabled = false;
    }
  });
  return button;
}

function createOnboardingApplicationCard(target: Application): HTMLElement {
  const card = document.createElement('article');
  card.className = 'onboarding-application-card';
  const selected = selectedApplicationIDs.has(target.id);
  card.classList.toggle('selected', selected);

  const icon = document.createElement('div');
  icon.className = `application-icon ${target.id}`;
  icon.textContent = symbols[target.id] ?? target.name.slice(0, 2);
  icon.setAttribute('aria-hidden', 'true');

  const information = document.createElement('div');
  information.className = 'application-info';
  const title = document.createElement('h3');
  title.textContent = target.name;
  const description = document.createElement('p');
  description.textContent = target.description;
  const metadata = document.createElement('div');
  metadata.className = 'metadata';
  const notice = legalNotices.find((candidate) => candidate.id === target.id);
  metadata.textContent = `Container próprio${notice?.license ? ` · ${notice.license}` : ''}`;
  if (target.dependencies.length > 0) {
    const dependencyNames = target.dependencies.map(
      (id) => availableApplications.find((candidate) => candidate.id === id)?.name ?? id,
    );
    metadata.textContent += ` · Inclui ${dependencyNames.join(', ')}`;
  }
  information.append(title, description, metadata);

  const actions = document.createElement('div');
  actions.className = 'onboarding-application-actions';
  if (notice?.links.some((link) => link.kind === 'license')) {
    const licenseButton = document.createElement('button');
    licenseButton.type = 'button';
    licenseButton.className = 'onboarding-license-button';
    licenseButton.textContent = 'Licença';
    licenseButton.addEventListener('click', () => void OpenLegalLink(target.id, 'license'));
    actions.append(licenseButton);
  }
  const selectButton = document.createElement('button');
  selectButton.type = 'button';
  selectButton.className = 'select-button';
  selectButton.textContent = selected ? 'Selecionado' : 'Selecionar';
  selectButton.setAttribute('aria-pressed', String(selected));
  selectButton.disabled = selectionSaving;
  selectButton.addEventListener('click', async () => {
    if (selectionSaving) return;
    const previousSelection = new Set(selectedApplicationIDs);
    if (selectedApplicationIDs.has(target.id)) {
      selectedApplicationIDs.delete(target.id);
    } else {
      selectedApplicationIDs.add(target.id);
    }
    selectionSaving = true;
    renderApplications();
    try {
      applySetupStatus(await SaveApplicationSelection([...selectedApplicationIDs]));
    } catch {
      selectedApplicationIDs = previousSelection;
      if (onboardingInstallationResult) {
        onboardingInstallationResult.textContent = 'Não foi possível salvar a seleção.';
        onboardingInstallationResult.classList.add('error');
      }
    } finally {
      selectionSaving = false;
      renderApplications();
    }
  });
  actions.append(selectButton);
  card.append(icon, information, actions);
  return card;
}

function renderApplications(): void {
  applicationsElement?.replaceChildren(...availableApplications.map(createApplicationCard));
  onboardingApplicationList?.replaceChildren(
    ...availableApplications.map(createOnboardingApplicationCard),
  );
  if (onboardingCatalogCount) {
    onboardingCatalogCount.textContent = `${availableApplications.length} disponíveis`;
  }
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
    renderApplications();
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
    updateJellyfinLANControl();
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

async function loadJellyfinNetwork(): Promise<void> {
  try {
    jellyfinNetwork = await GetJellyfinNetworkStatus();
    renderApplications();
  } catch {
    jellyfinNetwork = undefined;
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
    if (onboardingStorageTitle) onboardingStorageTitle.textContent = 'Pasta selecionada';
    if (onboardingStorageDescription) {
      onboardingStorageDescription.textContent =
        'O Corsarr verificará novamente esta pasta antes de continuar.';
    }
    if (onboardingStoragePath) onboardingStoragePath.textContent = status.storagePath;
    if (onboardingStorageBadge) {
      onboardingStorageBadge.textContent = 'Selecionada';
      onboardingStorageBadge.className = 'runtime-badge ready';
    }
    if (onboardingChooseStorage) onboardingChooseStorage.textContent = 'Trocar pasta';
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
  if (startAtLoginSetting) startAtLoginSetting.hidden = !status.startAtLoginSupported;
  if (startAtLoginCheckbox) {
    startAtLoginCheckbox.checked = status.startAtLogin;
    startAtLoginCheckbox.disabled = false;
  }
  if (openLoginSettingsButton) {
    openLoginSettingsButton.hidden = !status.startAtLoginRequiresApproval;
  }
  updateJellyfinLANControl();
  updateInstallAuthority();
  syncOnboarding(status);
}

function updateJellyfinLANControl(): void {
  const jellyfinSelected = setupStatus?.applications.includes('jellyfin') ?? false;
  const jellyfinState = managedStatuses.get('jellyfin')?.state ?? 'not_installed';
  const canChange = jellyfinState === 'not_installed';
  if (jellyfinLANSetting) jellyfinLANSetting.hidden = !jellyfinSelected;
  if (jellyfinLANCheckbox) {
    jellyfinLANCheckbox.checked = setupStatus?.jellyfinLanEnabled ?? false;
    jellyfinLANCheckbox.disabled = !canChange;
    jellyfinLANCheckbox.title = canChange
      ? ''
      : 'Remova o container do Jellyfin antes de alterar o acesso à rede. As configurações e a mídia serão preservadas.';
  }
}

function updateInstallAuthority(): void {
  if (!installApplicationsButton) return;
  installApplicationsButton.disabled =
    !setupStatus?.canPrepare || !acceptTermsCheckbox?.checked || !currentHostReady;
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
      'O Corsarr pode preparar os componentes necessários após sua autorização explícita.',
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
    currentHostReady = environment.host.ready;
    const runtimeMessage = runtimeMessages[environment.runtime.state] ?? runtimeMessages.error;

    if (platformElement) {
      platformElement.textContent = platformNames[environment.platform] ?? environment.platform;
    }
    if (architectureElement) {
      architectureElement.textContent =
        architectureNames[environment.architecture] ?? environment.architecture;
    }
    if (environmentTitleElement) {
      environmentTitleElement.textContent = currentHostReady
        ? runtimeMessage.title
        : 'Este computador precisa de atenção';
    }
    if (environmentDescriptionElement) {
      environmentDescriptionElement.textContent = currentHostReady
        ? runtimeMessage.description
        : environment.host.issues.join(' ');
    }
    if (environmentBadgeElement) {
      environmentBadgeElement.textContent = currentHostReady ? runtimeMessage.badge : 'Requisitos';
      environmentBadgeElement.className = `runtime-badge ${currentHostReady ? environment.runtime.state : 'error'}`;
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
      if (environment.host.osVersion) details.push(`macOS: ${environment.host.osVersion}`);
      if (environment.host.memoryBytes) {
        details.push(`Memória: ${(environment.host.memoryBytes / 1024 ** 3).toFixed(1)} GiB`);
      }
      if (environment.host.freeDiskBytes) {
        details.push(
          `Espaço livre: ${(environment.host.freeDiskBytes / 1024 ** 3).toFixed(1)} GiB`,
        );
      }
      environmentTechnicalElement.textContent = details.join('\n');
      if (onboardingEnvironmentTechnical) {
        onboardingEnvironmentTechnical.textContent = details.join('\n');
      }
    }
    if (prepareRuntimeButton) {
      prepareRuntimeButton.hidden = environment.runtime.state === 'ready';
      prepareRuntimeButton.disabled = !currentHostReady;
      prepareRuntimeButton.textContent =
        environment.runtime.state === 'stopped' ? 'Iniciar ambiente' : 'Preparar computador';
    }
    if (onboardingEnvironmentTitle) {
      onboardingEnvironmentTitle.textContent = currentHostReady
        ? runtimeMessage.title
        : 'Este computador precisa de atenção';
    }
    if (onboardingEnvironmentDescription) {
      onboardingEnvironmentDescription.textContent = currentHostReady
        ? runtimeMessage.description
        : environment.host.issues.join(' ');
    }
    if (onboardingEnvironmentBadge) {
      onboardingEnvironmentBadge.textContent = currentHostReady
        ? runtimeMessage.badge
        : 'Requisitos';
      onboardingEnvironmentBadge.className = `runtime-badge ${currentHostReady ? environment.runtime.state : 'error'}`;
    }
    if (onboardingPrepareRuntime) {
      onboardingPrepareRuntime.hidden = environment.runtime.state === 'ready';
      onboardingPrepareRuntime.disabled = !currentHostReady;
      onboardingPrepareRuntime.textContent =
        environment.runtime.state === 'stopped' ? 'Iniciar ambiente' : 'Preparar computador';
    }
    if (onboardingEnvironmentNext) {
      onboardingEnvironmentNext.disabled =
        !currentHostReady || environment.runtime.state !== 'ready';
    }
    updateInstallAuthority();
  } catch {
    currentRuntimeState = 'error';
    currentHostReady = false;
    if (environmentTitleElement) environmentTitleElement.textContent = 'Não foi possível verificar';
    if (environmentDescriptionElement) {
      environmentDescriptionElement.textContent = 'Tente novamente em alguns instantes.';
    }
    if (environmentBadgeElement) {
      environmentBadgeElement.textContent = 'Atenção';
      environmentBadgeElement.className = 'runtime-badge error';
    }
    if (onboardingEnvironmentTitle) {
      onboardingEnvironmentTitle.textContent = 'Não foi possível verificar';
    }
    if (onboardingEnvironmentDescription) {
      onboardingEnvironmentDescription.textContent = 'Tente novamente em alguns instantes.';
    }
    if (onboardingEnvironmentBadge) {
      onboardingEnvironmentBadge.textContent = 'Atenção';
      onboardingEnvironmentBadge.className = 'runtime-badge error';
    }
    if (onboardingEnvironmentNext) onboardingEnvironmentNext.disabled = true;
  } finally {
    refreshEnvironmentButton?.removeAttribute('disabled');
  }
}

refreshEnvironmentButton?.addEventListener('click', () => void loadEnvironment());

async function prepareRuntime(): Promise<void> {
  if (!prepareRuntimeButton) return;
  if (!currentHostReady) {
    if (installationResultElement) {
      installationResultElement.textContent =
        'Resolva os requisitos indicados neste computador antes de preparar o ambiente.';
      installationResultElement.classList.add('error');
    }
    return;
  }
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

async function chooseStorage(activeButton = chooseStorageButton): Promise<void> {
  if (!activeButton) return;
  activeButton.disabled = true;
  activeButton.textContent = 'Verificando…';

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
      if (onboardingStorageTitle) onboardingStorageTitle.textContent = 'Pasta pronta';
      if (onboardingStorageDescription) {
        onboardingStorageDescription.textContent = storage.hardlinks
          ? 'A pasta é gravável e permite organizar arquivos sem duplicação.'
          : 'A pasta é gravável; algumas importações poderão copiar arquivos.';
      }
      if (onboardingStoragePath) onboardingStoragePath.textContent = storage.path;
      if (onboardingStorageFacts) {
        onboardingStorageFacts.textContent = `${formatAvailableSpace(storage.availableBytes ?? 0)} · ${storage.hardlinks ? 'Hardlinks disponíveis' : 'Sem hardlinks'}`;
      }
      if (onboardingStorageBadge) {
        onboardingStorageBadge.textContent = storage.hardlinks ? 'Pronto' : 'Compatível';
        onboardingStorageBadge.className = `runtime-badge ${storage.hardlinks ? 'ready' : 'stopped'}`;
      }
      activeButton.textContent = 'Trocar pasta';
      return;
    }

    if (storageTitleElement) storageTitleElement.textContent = 'Esta pasta não pode ser usada';
    if (storageDescriptionElement) {
      storageDescriptionElement.textContent =
        storage.availableBytes !== undefined &&
        storage.requiredBytes !== undefined &&
        storage.availableBytes < storage.requiredBytes
          ? `Há apenas ${formatAvailableSpace(storage.availableBytes).replace(' disponíveis', '')}. O Corsarr precisa de pelo menos ${formatAvailableSpace(storage.requiredBytes).replace(' disponíveis', ' livres')}.`
          : (storage.technicalDetail ??
            'Escolha uma pasta existente com permissão de escrita e espaço disponível verificável.');
    }
    if (storagePathElement) storagePathElement.textContent = storage.path;
    if (storageFactsElement) storageFactsElement.textContent = '';
    if (storageBadgeElement) {
      storageBadgeElement.textContent = 'Atenção';
      storageBadgeElement.className = 'runtime-badge error';
    }
    if (onboardingStorageTitle)
      onboardingStorageTitle.textContent = 'Esta pasta não pode ser usada';
    if (onboardingStorageDescription) {
      onboardingStorageDescription.textContent =
        storage.technicalDetail ??
        'Escolha outra pasta com permissão de escrita e espaço disponível.';
    }
    if (onboardingStoragePath) onboardingStoragePath.textContent = storage.path;
    if (onboardingStorageFacts) onboardingStorageFacts.textContent = '';
    if (onboardingStorageBadge) {
      onboardingStorageBadge.textContent = 'Atenção';
      onboardingStorageBadge.className = 'runtime-badge error';
    }
  } catch {
    if (storageTitleElement) storageTitleElement.textContent = 'Não foi possível verificar a pasta';
    if (storageDescriptionElement) storageDescriptionElement.textContent = 'Tente novamente.';
    if (storageBadgeElement) {
      storageBadgeElement.textContent = 'Atenção';
      storageBadgeElement.className = 'runtime-badge error';
    }
    if (onboardingStorageTitle) onboardingStorageTitle.textContent = 'Não foi possível verificar';
    if (onboardingStorageDescription) onboardingStorageDescription.textContent = 'Tente novamente.';
    if (onboardingStorageBadge) {
      onboardingStorageBadge.textContent = 'Atenção';
      onboardingStorageBadge.className = 'runtime-badge error';
    }
  } finally {
    activeButton.disabled = false;
    if (activeButton.textContent === 'Verificando…') {
      activeButton.textContent = 'Escolher pasta';
    }
  }
}

chooseStorageButton?.addEventListener('click', () => void chooseStorage(chooseStorageButton));

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

startAtLoginCheckbox?.addEventListener('change', async () => {
  if (!startAtLoginCheckbox) return;
  const requested = startAtLoginCheckbox.checked;
  startAtLoginCheckbox.disabled = true;
  try {
    const status = await SetStartAtLogin(requested);
    applySetupStatus(status);
    if (messageElement) {
      messageElement.textContent = status.startAtLoginRequiresApproval
        ? 'O macOS precisa da sua aprovação em Ajustes do Sistema para iniciar o Corsarr no login.'
        : requested
          ? 'O Corsarr iniciará seus serviços quando você entrar neste Mac.'
          : 'O Corsarr não iniciará automaticamente no próximo login.';
      messageElement.classList.remove('error');
    }
  } catch {
    applySetupStatus(await GetSetupStatus());
    if (messageElement) {
      messageElement.textContent = 'Não foi possível alterar o início automático.';
      messageElement.classList.add('error');
    }
  }
});

jellyfinLANCheckbox?.addEventListener('change', async () => {
  if (!jellyfinLANCheckbox) return;
  const requested = jellyfinLANCheckbox.checked;
  if (
    requested &&
    !window.confirm(
      'Permitir acesso ao Jellyfin nesta rede local? TVs, celulares e outros aparelhos conectados à mesma rede poderão acessar a tela do Jellyfin.',
    )
  ) {
    jellyfinLANCheckbox.checked = false;
    return;
  }
  jellyfinLANCheckbox.disabled = true;
  try {
    applySetupStatus(await SetJellyfinLAN(requested));
    if (messageElement) {
      messageElement.textContent = requested
        ? 'O Jellyfin ficará disponível para os aparelhos desta rede após a instalação.'
        : 'O Jellyfin ficará acessível somente neste computador.';
      messageElement.classList.remove('error');
    }
  } catch {
    applySetupStatus(await GetSetupStatus());
    if (messageElement) {
      messageElement.textContent = 'Não foi possível alterar o acesso do Jellyfin à rede.';
      messageElement.classList.add('error');
    }
  }
});

openLoginSettingsButton?.addEventListener('click', async () => {
  openLoginSettingsButton.disabled = true;
  try {
    await OpenStartAtLoginSettings();
  } finally {
    openLoginSettingsButton.disabled = false;
  }
});

async function installApplications(): Promise<void> {
  if (!installApplicationsButton || !setupStatus?.canPrepare || !acceptTermsCheckbox?.checked) {
    return;
  }
  if (
    currentRuntimeState === 'unavailable' &&
    !window.confirm(
      'O Corsarr precisa preparar este computador antes de instalar os aplicativos. O Docker Desktop 4.86.0 será baixado da fonte oficial, verificado e o macOS pedirá sua autorização. Continuar?',
    )
  ) {
    return;
  }

  installApplicationsButton.disabled = true;
  installApplicationsButton.textContent = 'Instalando…';
  if (installationResultElement) {
    installationResultElement.textContent =
      'Preparando o ambiente e baixando os aplicativos. Isso pode levar alguns minutos.';
    installationResultElement.classList.remove('error');
  }
  renderOperationIssue();

  try {
    applySetupStatus(await AcceptCurrentTerms());
    const result = await InstallSelectedApplications();
    if (result.complete) {
      renderOperationIssue();
      if (installationResultElement) {
        installationResultElement.textContent = `${result.items.length} aplicativos instalados e iniciados.`;
      }
      installApplicationsButton.textContent = 'Aplicativos instalados';
      await Promise.all([
        loadApplicationStatuses(),
        loadApplicationDataStatuses(),
        loadJellyfinAccess(),
        loadJellyfinNetwork(),
        loadQBittorrentAccess(),
      ]);
    } else {
      const failed = result.items.find((item) => item.failed);
      renderOperationIssue(failed?.issue);
      if (installationResultElement) {
        installationResultElement.textContent = failed
          ? `${failed.issue?.summary ?? `A instalação de ${failed.applicationId} não terminou.`} ${failed.issue?.nextAction ?? 'Tente novamente.'}`
          : 'A instalação não terminou. Tente novamente.';
        installationResultElement.classList.add('error');
      }
      installApplicationsButton.textContent = 'Tentar novamente';
    }
  } catch {
    renderOperationIssue();
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

for (const backButton of document.querySelectorAll<HTMLButtonElement>('.onboarding-back')) {
  backButton.addEventListener('click', () => {
    showOnboardingStep(backButton.dataset.onboardingStep as OnboardingStep);
  });
}

onboardingStartButton?.addEventListener('click', async () => {
  onboardingStartButton.disabled = true;
  try {
    if (onboardingHasAdvancedPast('splash')) {
      if (setupStatus) syncOnboarding(setupStatus);
      return;
    }
    applySetupStatus(await AdvanceOnboarding());
  } finally {
    onboardingStartButton.disabled = false;
  }
});

onboardingTermsCheckbox?.addEventListener('change', () => {
  if (onboardingPermissionsNext) {
    onboardingPermissionsNext.disabled = !onboardingTermsCheckbox.checked;
  }
});

onboardingOpenDockerTerms?.addEventListener('click', async () => {
  onboardingOpenDockerTerms.disabled = true;
  try {
    await OpenLegalLink('runtime-docker', 'license');
  } finally {
    onboardingOpenDockerTerms.disabled = false;
  }
});

onboardingPermissionsNext?.addEventListener('click', async () => {
  if (!onboardingTermsCheckbox?.checked || !onboardingPermissionsNext) return;
  const resumeExistingProgress = onboardingHasAdvancedPast('permissions');
  const startAtLogin = onboardingStartLoginCheckbox?.checked ?? false;
  onboardingPermissionsNext.disabled = true;
  if (onboardingPermissionsMessage) {
    onboardingPermissionsMessage.textContent = 'Salvando suas escolhas…';
    onboardingPermissionsMessage.classList.remove('error');
  }
  try {
    applySetupStatus(await AcceptCurrentTerms());
    if (setupStatus?.startAtLoginSupported) {
      applySetupStatus(await SetStartAtLogin(startAtLogin));
    }
    if (!resumeExistingProgress) applySetupStatus(await AdvanceOnboarding());
    if (onboardingPermissionsMessage) onboardingPermissionsMessage.textContent = '';
  } catch {
    if (onboardingPermissionsMessage) {
      onboardingPermissionsMessage.textContent =
        'Não foi possível salvar a autorização. Nenhum componente foi instalado.';
      onboardingPermissionsMessage.classList.add('error');
    }
  } finally {
    onboardingPermissionsNext.disabled = !onboardingTermsCheckbox.checked;
  }
});

onboardingPrepareRuntime?.addEventListener('click', async () => {
  if (!onboardingPrepareRuntime || !currentHostReady) return;
  if (
    currentRuntimeState === 'unavailable' &&
    !window.confirm(
      'O Corsarr baixará o Docker Desktop 4.86.0 da fonte oficial, verificará sua integridade e solicitará a autorização do macOS. Continuar?',
    )
  ) {
    return;
  }
  onboardingPrepareRuntime.disabled = true;
  onboardingPrepareRuntime.textContent =
    currentRuntimeState === 'unavailable' ? 'Instalando…' : 'Iniciando…';
  if (onboardingEnvironmentMessage) {
    onboardingEnvironmentMessage.textContent =
      'Preparando os componentes necessários. O macOS poderá solicitar sua senha.';
    onboardingEnvironmentMessage.classList.remove('error');
  }
  try {
    const result = await PrepareRuntime();
    if (!result.ready) throw new Error('runtime not ready');
    await loadEnvironment();
    if (onboardingEnvironmentMessage) {
      onboardingEnvironmentMessage.textContent = result.installed
        ? 'Docker Desktop instalado e pronto.'
        : 'Ambiente iniciado e pronto.';
    }
  } catch {
    if (onboardingEnvironmentMessage) {
      onboardingEnvironmentMessage.textContent =
        'A preparação não terminou. Tente novamente ou consulte os detalhes técnicos.';
      onboardingEnvironmentMessage.classList.add('error');
    }
    await loadEnvironment();
  }
});

onboardingEnvironmentNext?.addEventListener('click', async () => {
  if (!onboardingEnvironmentNext || onboardingEnvironmentNext.disabled) return;
  onboardingEnvironmentNext.disabled = true;
  try {
    if (onboardingHasAdvancedPast('environment')) {
      if (setupStatus) syncOnboarding(setupStatus);
      return;
    }
    applySetupStatus(await AdvanceOnboarding());
  } catch {
    if (onboardingEnvironmentMessage) {
      onboardingEnvironmentMessage.textContent =
        'Confirme que o ambiente está pronto antes de continuar.';
      onboardingEnvironmentMessage.classList.add('error');
    }
  } finally {
    onboardingEnvironmentNext.disabled = currentRuntimeState !== 'ready' || !currentHostReady;
  }
});

onboardingChooseStorage?.addEventListener(
  'click',
  () => void chooseStorage(onboardingChooseStorage),
);

onboardingStorageNext?.addEventListener('click', async () => {
  if (!onboardingStorageNext || !setupStatus?.storagePath) return;
  onboardingStorageNext.disabled = true;
  if (onboardingStorageMessage) {
    onboardingStorageMessage.textContent = 'Verificando a pasta novamente…';
    onboardingStorageMessage.classList.remove('error');
  }
  try {
    if (onboardingHasAdvancedPast('storage')) {
      syncOnboarding(setupStatus);
      return;
    }
    applySetupStatus(await AdvanceOnboarding());
    if (onboardingStorageMessage) onboardingStorageMessage.textContent = '';
  } catch {
    if (onboardingStorageMessage) {
      onboardingStorageMessage.textContent =
        'A pasta não está mais disponível ou não possui espaço suficiente. Escolha outra pasta.';
      onboardingStorageMessage.classList.add('error');
    }
  } finally {
    onboardingStorageNext.disabled = !setupStatus?.storagePath;
  }
});

onboardingRecommended?.addEventListener('click', async () => {
  if (!onboardingRecommended || selectionSaving) return;
  onboardingRecommended.disabled = true;
  selectionSaving = true;
  renderApplications();
  try {
    applySetupStatus(await SelectRecommendedApplications());
    if (onboardingInstallationResult) {
      onboardingInstallationResult.textContent =
        'Configuração recomendada selecionada. Você ainda pode personalizá-la.';
      onboardingInstallationResult.classList.remove('error');
    }
  } catch {
    if (onboardingInstallationResult) {
      onboardingInstallationResult.textContent =
        'Não foi possível selecionar a configuração recomendada.';
      onboardingInstallationResult.classList.add('error');
    }
  } finally {
    selectionSaving = false;
    onboardingRecommended.disabled = false;
    renderApplications();
  }
});

onboardingJellyfinLANCheckbox?.addEventListener('change', async () => {
  if (!onboardingJellyfinLANCheckbox) return;
  const enabled = onboardingJellyfinLANCheckbox.checked;
  if (
    enabled &&
    !window.confirm('Permitir que TVs e celulares desta rede local acessem o Jellyfin?')
  ) {
    onboardingJellyfinLANCheckbox.checked = false;
    return;
  }
  onboardingJellyfinLANCheckbox.disabled = true;
  try {
    applySetupStatus(await SetJellyfinLAN(enabled));
  } catch {
    applySetupStatus(await GetSetupStatus());
  } finally {
    onboardingJellyfinLANCheckbox.disabled = false;
  }
});

onboardingInstallButton?.addEventListener('click', async () => {
  if (!onboardingInstallButton || !setupStatus?.canInstall) return;
  onboardingInstallButton.disabled = true;
  onboardingInstallButton.textContent = 'Instalando…';
  renderOnboardingIssue();
  if (onboardingInstallationResult) {
    onboardingInstallationResult.textContent =
      'Baixando e configurando os aplicativos. Isso pode levar alguns minutos.';
    onboardingInstallationResult.classList.remove('error');
  }
  try {
    const result = await InstallSelectedApplications();
    if (!result.complete) {
      const failed = result.items.find((item) => item.failed);
      renderOnboardingIssue(failed?.issue);
      if (onboardingInstallationResult) {
        onboardingInstallationResult.textContent = `${failed?.issue?.summary ?? 'A instalação não terminou.'} ${failed?.issue?.nextAction ?? 'Tente novamente.'}`;
        onboardingInstallationResult.classList.add('error');
      }
      onboardingInstallButton.textContent = 'Tentar novamente';
      return;
    }
    applySetupStatus(await GetSetupStatus());
    await Promise.all([
      loadApplicationStatuses(),
      loadApplicationDataStatuses(),
      loadJellyfinAccess(),
      loadJellyfinNetwork(),
      loadQBittorrentAccess(),
    ]);
    if (messageElement) {
      messageElement.textContent = `${result.items.length} aplicativos instalados. A configuração inicial foi concluída.`;
      messageElement.classList.remove('error');
    }
  } catch {
    if (onboardingInstallationResult) {
      onboardingInstallationResult.textContent =
        'Não foi possível concluir a instalação. Suas escolhas foram preservadas.';
      onboardingInstallationResult.classList.add('error');
    }
    onboardingInstallButton.textContent = 'Tentar novamente';
  } finally {
    if (!setupStatus?.onboardingCompleted) {
      onboardingInstallButton.disabled = !setupStatus?.canInstall;
    }
  }
});

async function loadInitialState(): Promise<void> {
  await Promise.all([loadEnvironment(), loadSetup()]);
  await Promise.all([loadApplications(), loadLegalNotices()]);
  await Promise.all([
    loadApplicationStatuses(),
    loadApplicationDataStatuses(),
    loadJellyfinAccess(),
    loadJellyfinNetwork(),
    loadQBittorrentAccess(),
  ]);
}

EventsOn('corsarr:background-recovery-complete', () => {
  void Promise.all([
    loadEnvironment(),
    loadApplicationStatuses(),
    loadJellyfinAccess(),
    loadJellyfinNetwork(),
    loadQBittorrentAccess(),
  ]);
});

EventsOn('corsarr:installation-progress', (progress: InstallationProgress) => {
  const applicationName =
    availableApplications.find((application) => application.id === progress.applicationId)?.name ??
    'aplicativo';
  const stageMessages: Record<InstallationProgress['stage'], string> = {
    installing: `Baixando e iniciando ${applicationName}`,
    provisioning: `Configurando ${applicationName}`,
    ready: `${applicationName} está pronto`,
    failed: `${applicationName} precisa de atenção`,
  };
  if (installationResultElement) {
    installationResultElement.textContent = `${stageMessages[progress.stage]} (${progress.position} de ${progress.total}).`;
    installationResultElement.classList.toggle('error', progress.stage === 'failed');
  }
  if (onboardingInstallationResult && !onboardingElement?.hidden) {
    onboardingInstallationResult.textContent = `${stageMessages[progress.stage]} (${progress.position} de ${progress.total}).`;
    onboardingInstallationResult.classList.toggle('error', progress.stage === 'failed');
  }
  if (installApplicationsButton && progress.stage !== 'failed') {
    installApplicationsButton.textContent = `${progress.position} de ${progress.total}`;
  }
  if (onboardingInstallButton && progress.stage !== 'failed') {
    onboardingInstallButton.textContent =
      progress.stage === 'ready' && progress.position === progress.total
        ? 'Finalizando…'
        : `${progress.position} de ${progress.total}`;
  }
});

void loadInitialState();
