import './style.css';
import {
  AcceptCurrentTerms,
  AdvanceOnboarding,
  ArchiveApplicationData,
  ChooseStorageLocation,
  CopyARRPassword,
  CopyJellyfinNetworkURL,
  CopyJellyfinPassword,
  CopyLastInstallationSupportReport,
  CopyLazyLibrarianPassword,
  CopyQBittorrentPassword,
  ExportDiagnostics,
  GetApplicationDataStatuses,
  GetApplicationStatuses,
  GetARRAccessStatuses,
  GetEnvironmentStatus,
  GetJellyfinAccessStatus,
  GetJellyfinNetworkStatus,
  GetLazyLibrarianAccessStatus,
  GetProductInfo,
  GetQBittorrentAccessStatus,
  GetSetupStatus,
  InstallSelectedApplications,
  ListApplications,
  ListLegalNotices,
  ListQualityProfilePresets,
  OpenApplication,
  OpenLegalLink,
  OpenStartAtLoginSettings,
  PrepareRuntime,
  PrepareStorageLayout,
  RemoveApplication,
  RestartApplication,
  SaveApplicationSelection,
  SaveQualityProfilePreset,
  SelectRecommendedApplications,
  SetJellyfinLAN,
  SetLanguagePreference,
  SetStartAtLogin,
  StartApplication,
  StopApplication,
  UpdateApplication,
} from '../wailsjs/go/main/App';
import type { application, legal, main, quality, storage } from '../wailsjs/go/models';
import { EventsOn } from '../wailsjs/runtime/runtime';
import {
  missingSelectedIntegrations,
  selectApplicationWithIntegrations,
  toggleApplicationSelection,
} from './application-selection';
import { runningServicesSummary, sortApplicationsByInstallation } from './dashboard-applications';
import {
  currentLocale,
  detectLocale,
  initializeLocalization,
  languageStorageKey,
  normalizeLocale,
  type TranslationKey,
  translate as t,
} from './i18n';
import {
  applyInstallationProgress,
  createInstallationProgress,
  type InstallationProgressEvent,
  type InstallationProgressItem,
  type TrackedInstallationStage,
} from './installation-progress';
import { onboardingStepTotal, shouldManageQualityProfile } from './quality-profile-selection';

type Application = application.ApplicationSummary;
type ManagedStatus = application.ManagedApplicationStatus;
type DataStatus = storage.ApplicationDataStatus;
type LegalNotice = legal.Notice;
type QualityPreset = quality.Preset;
const detectedLocale = detectLocale(
  window.localStorage.getItem(languageStorageKey),
  navigator.languages,
);
initializeLocalization(detectedLocale);
const root = document.querySelector<HTMLDivElement>('#app');

if (!root) {
  throw new Error('Corsarr root element was not found.');
}

root.innerHTML = [
  '<section id="onboarding" class="onboarding" hidden aria-label="Configuração inicial do Corsarr">',
  '  <div id="onboarding-notification" class="onboarding-notification" role="alert" aria-live="assertive" hidden><span class="onboarding-notification-icon" aria-hidden="true">!</span><p id="onboarding-notification-message"></p><button id="onboarding-notification-dismiss" type="button" aria-label="Fechar notificação">Fechar</button></div>',
  '  <div class="onboarding-frame">',
  '    <header class="onboarding-header">',
  '      <div class="onboarding-brand"><span class="brand-mark" aria-hidden="true">C</span><span><strong>Corsarr</strong><small>Configuração inicial</small></span></div>',
  `      <label class="language-control"><span>${t('language.label')}</span><select class="language-select" aria-label="${t('language.label')}"><option value="en">${t('language.en')}</option><option value="es">${t('language.es')}</option><option value="pt-BR">${t('language.pt-BR')}</option><option value="it">${t('language.it')}</option></select></label>`,
  '      <div id="onboarding-progress" class="onboarding-progress" hidden><span class="active"></span><span></span><span></span><span></span><span></span><small id="onboarding-progress-label">Etapa 1 de 4</small></div>',
  '    </header>',
  '    <article id="onboarding-splash" class="onboarding-step onboarding-splash">',
  '      <div class="onboarding-splash-copy"><p class="eyebrow">BEM-VINDO AO CORSARR</p><h1>Seu servidor de mídia,<br><em>sem complicação.</em></h1><p>Vamos preparar este computador juntos. Você entenderá cada escolha antes que o Corsarr instale qualquer componente.</p><button id="onboarding-start" class="onboarding-primary" type="button">Começar configuração</button></div>',
  '      <div class="radar onboarding-radar" aria-hidden="true"><span></span><span></span><i></i><b>C</b></div>',
  '    </article>',
  '    <article id="onboarding-permissions" class="onboarding-step" hidden>',
  '      <div class="onboarding-step-copy"><p class="eyebrow">ETAPA 1 · AUTORIZAÇÃO</p><h1>Você mantém o controle.</h1><p>O Corsarr usará o Docker Desktop para executar os aplicativos em serviços isolados. Ele só instalará componentes, baixará imagens e criará serviços depois da sua autorização.</p><div class="onboarding-explanation"><strong>O que será autorizado</strong><ul><li>Usar ou instalar o Docker Desktop neste Mac.</li><li>Baixar somente imagens aprovadas e identificadas por versão.</li><li>Criar pastas e serviços apenas dentro da configuração do Corsarr.</li></ul></div><label class="onboarding-check"><input id="onboarding-terms" type="checkbox"><span>Li e autorizo o uso do Docker Desktop e dos aplicativos selecionados. Entendo que cada componente mantém sua própria licença.</span></label><button id="onboarding-open-docker-terms" class="onboarding-link" type="button">Abrir termos oficiais do Docker Desktop</button><label id="onboarding-login-setting" class="onboarding-check"><input id="onboarding-start-login" type="checkbox"><span>Iniciar meus serviços automaticamente quando eu entrar neste Mac.</span></label></div>',
  '      <footer class="onboarding-actions"><button class="onboarding-back" type="button" data-onboarding-step="splash">Voltar</button><button id="onboarding-permissions-next" class="onboarding-primary" type="button" disabled>Autorizar e continuar</button></footer>',
  '    </article>',
  '    <article id="onboarding-environment" class="onboarding-step" hidden>',
  '      <div class="onboarding-step-copy"><p class="eyebrow">ETAPA 2 · AMBIENTE</p><h1>Preparando este computador.</h1><p>O Docker mantém cada aplicativo separado e permite que o Corsarr cuide de instalação, atualização e reinício por você.</p><div class="onboarding-diagnostic"><span class="environment-icon" aria-hidden="true">◎</span><div><small>DIAGNÓSTICO</small><strong id="onboarding-environment-title">Verificando…</strong><p id="onboarding-environment-description">Aguarde enquanto conferimos este Mac.</p><details><summary>Detalhes técnicos</summary><code id="onboarding-environment-technical"></code></details></div><span id="onboarding-environment-badge" class="runtime-badge checking">Verificando</span></div><p id="onboarding-environment-message" class="onboarding-message"></p></div>',
  '      <footer class="onboarding-actions"><button class="onboarding-back" type="button" data-onboarding-step="permissions">Voltar</button><div><button id="onboarding-prepare-runtime" class="secondary-button" type="button" hidden>Preparar computador</button><button id="onboarding-environment-next" class="onboarding-primary" type="button" disabled>Próximo</button></div></footer>',
  '    </article>',
  '    <article id="onboarding-storage" class="onboarding-step" hidden>',
  '      <div class="onboarding-step-copy"><p class="eyebrow">ETAPA 3 · ARMAZENAMENTO</p><h1>Onde sua mídia ficará?</h1><p>Dentro da pasta escolhida, o Corsarr criará uma estrutura clara para configurações, downloads e bibliotecas. Você poderá encontrá-la depois sem conhecer containers.</p><div class="onboarding-path-preview"><span class="storage-icon" aria-hidden="true">▱</span><div><strong id="onboarding-storage-title">Nenhuma pasta selecionada</strong><p id="onboarding-storage-description">Escolha uma pasta com pelo menos 10 GB disponíveis.</p><code id="onboarding-storage-path"></code><small id="onboarding-storage-facts"></small></div><span id="onboarding-storage-badge" class="runtime-badge checking">Pendente</span></div><div class="folder-tree" aria-label="Estrutura que será criada"><code>Corsarr/<br>├── config/ <span>configurações dos aplicativos</span><br>└── media/<br>&nbsp;&nbsp;&nbsp;├── downloads/ <span>arquivos baixados</span><br>&nbsp;&nbsp;&nbsp;└── library/ <span>filmes, séries, músicas e livros</span></code></div><p id="onboarding-storage-message" class="onboarding-message"></p></div>',
  '      <footer class="onboarding-actions"><button class="onboarding-back" type="button" data-onboarding-step="environment">Voltar</button><div><button id="onboarding-choose-storage" class="secondary-button" type="button">Escolher pasta</button><button id="onboarding-storage-next" class="onboarding-primary" type="button" disabled>Próximo</button></div></footer>',
  '    </article>',
  '    <article id="onboarding-applications" class="onboarding-step" hidden>',
  '      <div class="onboarding-step-copy onboarding-applications-copy"><p class="eyebrow">ETAPA 4 · APLICATIVOS</p><h1>Escolha o que deseja instalar.</h1><div id="onboarding-integration-guidance" class="onboarding-integration-benefit" role="status" aria-live="polite"><strong id="onboarding-integration-title">Mais automação, menos configuração</strong><p id="onboarding-integration-copy">Quando o Corsarr instala todos os aplicativos recomendados, ele pode conectar buscas e downloads para você. Ao escolher um aplicativo, também marcaremos as integrações recomendadas. Você pode desmarcar qualquer item se já usa seu próprio serviço.</p></div><div class="onboarding-catalog-heading"><span id="onboarding-catalog-count">Carregando…</span><button id="onboarding-recommended" class="secondary-button" type="button">Usar configuração recomendada</button></div><div id="onboarding-application-list" class="onboarding-application-list"></div><label id="onboarding-jellyfin-lan-setting" class="onboarding-check" hidden><input id="onboarding-jellyfin-lan" type="checkbox"><span>Permitir assistir no Jellyfin por TVs e aparelhos desta rede local.</span></label><p id="onboarding-application-message" class="onboarding-message" aria-live="polite"></p></div>',
  '      <footer class="onboarding-actions"><button class="onboarding-back" type="button" data-onboarding-step="storage">Voltar</button><button id="onboarding-install" class="onboarding-primary" type="button" disabled>Instalar aplicativos</button></footer>',
  '    </article>',
  '    <article id="onboarding-quality" class="onboarding-step" hidden>',
  '      <div class="onboarding-step-copy"><p class="eyebrow">ETAPA 5 · QUALIDADE</p><h1>Como você prefere sua mídia?</h1><p>O Corsarr usará recomendações versionadas do TRaSH Guides para configurar Radarr e Sonarr. Você poderá revisar e alterar esta escolha depois.</p><div id="onboarding-quality-list" class="quality-profile-list" role="radiogroup" aria-label="Perfil de qualidade"></div><div class="onboarding-explanation"><strong>Antes de concluir</strong><p id="onboarding-quality-summary">Escolha um perfil para continuar.</p></div><p id="onboarding-quality-result" class="onboarding-message" aria-live="polite"></p></div>',
  '      <footer class="onboarding-actions"><button class="onboarding-back" type="button" data-onboarding-step="applications">Voltar</button><button id="onboarding-quality-install" class="onboarding-primary" type="button" disabled>Instalar e aplicar perfil</button></footer>',
  '    </article>',
  '    <article id="onboarding-installation" class="onboarding-step onboarding-installation-step" hidden>',
  '      <div class="onboarding-step-copy onboarding-installation-copy"><p class="eyebrow">INSTALAÇÃO EM ANDAMENTO</p><h1 id="onboarding-installation-title" tabindex="-1">Preparando seu servidor.</h1><p>Você pode acompanhar cada aplicativo enquanto o Corsarr baixa, inicia e configura os serviços selecionados.</p><p id="onboarding-installation-result" class="onboarding-message" role="status" aria-live="polite">Verificando o ambiente e preparando a instalação…</p><section id="onboarding-installation-progress" class="installation-progress-panel" aria-label="Progresso da instalação"><header><span>PROGRESSO GERAL</span><strong id="onboarding-installation-progress-summary">Preparando</strong></header><div id="onboarding-installation-progress-track" class="installation-progress-track" role="progressbar" aria-label="Instalação dos aplicativos" aria-valuemin="0" aria-valuemax="100" aria-valuenow="0"><span id="onboarding-installation-progress-bar"></span></div><ol id="onboarding-installation-progress-list"></ol><div id="onboarding-installation-completion" class="installation-completion waiting"><span class="installation-progress-indicator" aria-hidden="true"></span><div><strong id="onboarding-installation-completion-title">Finalizando configuração</strong><small id="onboarding-installation-completion-status">Aguardando os aplicativos</small></div></div></section><details id="onboarding-operation-details" class="operation-details" hidden><summary>Detalhes técnicos</summary><code id="onboarding-operation-technical"></code></details></div>',
  '      <footer class="onboarding-actions onboarding-installation-actions"><strong class="onboarding-installation-warning">Não feche o Corsarr durante esta etapa.</strong><div><button id="onboarding-copy-support-report" class="secondary-button support-report-button" type="button" hidden>Copiar log de erros</button><button id="onboarding-installation-retry" class="onboarding-primary" type="button" hidden>Tentar novamente</button></div></footer>',
  '    </article>',
  '  </div>',
  '</section>',
  '<div id="dashboard-shell" class="shell" hidden>',
  '  <aside class="sidebar">',
  `    <a class="brand" href="#" aria-label="${t('dashboard.aria')}">`,
  '      <span class="brand-mark" aria-hidden="true">C</span>',
  '      <span><strong>Corsarr</strong><small>Desktop</small></span>',
  '    </a>',
  `    <nav aria-label="${t('nav.aria')}">`,
  `      <button id="show-home" class="nav-item active" type="button"><span aria-hidden="true">⌂</span>${t('nav.home')}</button>`,
  `      <button id="show-info" class="nav-item" type="button"><span aria-hidden="true">ⓘ</span>${t('nav.info')}</button>`,
  `      <button id="show-licenses" class="nav-item" type="button"><span aria-hidden="true">§</span>${t('nav.licenses')}</button>`,
  `      <button id="export-diagnostics" class="nav-item" type="button"><span aria-hidden="true">⇩</span>${t('nav.exportDiagnostics')}</button>`,
  '    </nav>',
  `    <label class="language-control sidebar-language"><span>${t('language.label')}</span><select class="language-select" aria-label="${t('language.label')}"><option value="en">${t('language.en')}</option><option value="es">${t('language.es')}</option><option value="pt-BR">${t('language.pt-BR')}</option><option value="it">${t('language.it')}</option></select></label>`,
  '  </aside>',
  '  <main class="content">',
  '    <div id="home-view">',
  '    <header class="topbar">',
  `      <div><p class="eyebrow">${t('dashboard.eyebrow')}</p><h1>${t('dashboard.greeting')}</h1></div>`,
  `      <div class="machine"><span class="machine-icon" aria-hidden="true">⌘</span><span><small id="platform-name">${t('dashboard.checkingComputer')}</small><strong id="architecture-name">${t('dashboard.wait')}</strong></span></div>`,
  '    </header>',
  '    <section class="hero" aria-labelledby="hero-title">',
  '      <div>',
  `        <span class="phase">${t('dashboard.localControl')}</span>`,
  `        <h2 id="hero-title">${t('dashboard.heroTitle')}<br><em>${t('dashboard.runningServices', { running: 0, installed: 0 })}</em></h2>`,
  `        <p>${t('dashboard.heroDescription')}</p>`,
  '      </div>',
  '      <div class="radar" aria-hidden="true"><span></span><span></span><i></i><b>C</b></div>',
  '    </section>',
  '    <section class="environment-card" aria-labelledby="environment-title">',
  '      <span class="environment-icon" aria-hidden="true">◎</span>',
  '      <div class="environment-copy">',
  `        <p class="eyebrow">${t('dashboard.thisComputer')}</p>`,
  `        <h2 id="environment-title">${t('dashboard.checkingEnvironment')}</h2>`,
  `        <p id="environment-description">${t('dashboard.environmentDescription')}</p>`,
  `        <details id="environment-details"><summary>${t('common.technicalDetails')}</summary><code id="environment-technical">${t('dashboard.runningDiagnostics')}</code></details>`,
  '      </div>',
  `      <span id="environment-badge" class="runtime-badge checking">${t('dashboard.checking')}</span>`,
  `      <button id="prepare-runtime" class="choose-storage-button" type="button" hidden>${t('dashboard.prepareComputer')}</button>`,
  `      <button id="refresh-environment" class="refresh-button" type="button" aria-label="${t('dashboard.refreshEnvironment')}">↻</button>`,
  '    </section>',
  '    <section class="storage-card" aria-labelledby="storage-title">',
  '      <span class="storage-icon" aria-hidden="true">▱</span>',
  '      <div class="storage-copy">',
  `        <p class="eyebrow">${t('dashboard.storage')}</p>`,
  `        <h2 id="storage-title">${t('dashboard.chooseMediaStorage')}</h2>`,
  `        <p id="storage-description">${t('dashboard.storageDescription')}</p>`,
  '        <p id="storage-path" class="storage-path"></p>',
  '        <p id="storage-facts" class="storage-facts"></p>',
  '      </div>',
  `      <span id="storage-badge" class="runtime-badge checking">${t('dashboard.notChecked')}</span>`,
  `      <button id="choose-storage" class="choose-storage-button" type="button">${t('dashboard.chooseFolder')}</button>`,
  '    </section>',
  '    <section class="section-heading">',
  `      <div><p class="eyebrow">${t('dashboard.applications')}</p><h2>${t('dashboard.yourApplications')}</h2></div>`,
  `      <p id="catalog-count" class="catalog-count">${t('common.loading')}</p>`,
  '    </section>',
  '    <div id="message" class="message" role="status" aria-live="polite"></div>',
  `    <section id="applications" class="applications" aria-label="${t('dashboard.availableAppsAria')}">`,
  '      <div class="loading-card"></div><div class="loading-card"></div><div class="loading-card"></div>',
  '    </section>',
  '    </div>',
  '    <section id="licenses-view" class="licenses-view" hidden aria-labelledby="licenses-title">',
  `      <header class="credits-header"><div><p class="eyebrow">${t('licenses.eyebrow')}</p><h1 id="licenses-title">${t('licenses.title')}</h1><p>${t('licenses.description')}</p></div><button id="licenses-back" class="secondary-button" type="button">${t('common.backHome')}</button></header>`,
  `      <p class="affiliation-note">${t('licenses.affiliation')}</p>`,
  '      <div id="legal-notices" class="legal-notices"><div class="loading-card"></div></div>',
  '    </section>',
  '    <section id="info-view" class="info-view" hidden aria-labelledby="info-title">',
  `      <header class="credits-header"><div><p class="eyebrow">${t('info.eyebrow')}</p><h1 id="info-title">${t('info.title')}</h1><p>${t('info.description')}</p></div><button id="info-back" class="secondary-button" type="button">${t('common.backHome')}</button></header>`,
  '      <div class="info-grid">',
  `        <article class="info-card info-card-primary"><span class="info-icon" aria-hidden="true">C</span><div><small>${t('info.application')}</small><h2>Corsarr Desktop</h2><p>${t('info.installedVersion')}</p></div><strong id="info-corsarr-version">${t('common.loading')}</strong></article>`,
  `        <article class="info-card"><small>${t('info.qualityPolicy')}</small><h2 id="info-quality-policy">${t('common.loading')}</h2><p>${t('info.qualityPolicyDescription')}</p></article>`,
  `        <article class="info-card"><small>${t('info.synchronization')}</small><h2><button id="info-open-recyclarr" class="info-external-link" type="button" aria-label="${t('info.openRecyclarr')}">Recyclarr <span id="info-recyclarr-version">—</span><span aria-hidden="true">↗</span></button></h2><p>${t('info.recyclarrDescription')}</p></article>`,
  `        <article class="info-card"><small>${t('info.recommendations')}</small><h2><button id="info-open-trash-guides" class="info-external-link" type="button" aria-label="${t('info.openTrash')}">TRaSH Guides <span aria-hidden="true">↗</span></button></h2><p>${t('info.trashSource')} <code id="info-trash-guides-commit">—</code>.</p></article>`,
  `        <article class="info-card"><small>${t('info.profileUpdates')}</small><h2 id="info-automatic-updates">${t('info.disabled')}</h2><p>${t('info.profileUpdatesDescription')}</p></article>`,
  '      </div>',
  '      <p id="info-message" class="message" role="status" aria-live="polite"></p>',
  '    </section>',
  '  </main>',
  '</div>',
].join('');

const applicationsElement = document.querySelector<HTMLElement>('#applications');
const heroTitleElement = document.querySelector<HTMLElement>('#hero-title');
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
const onboardingNotification = document.querySelector<HTMLElement>('#onboarding-notification');
const onboardingNotificationIcon = document.querySelector<HTMLElement>(
  '.onboarding-notification-icon',
);
const onboardingNotificationMessage = document.querySelector<HTMLElement>(
  '#onboarding-notification-message',
);
const onboardingNotificationDismiss = document.querySelector<HTMLButtonElement>(
  '#onboarding-notification-dismiss',
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
const onboardingIntegrationGuidance = document.querySelector<HTMLElement>(
  '#onboarding-integration-guidance',
);
const onboardingIntegrationTitle = document.querySelector<HTMLElement>(
  '#onboarding-integration-title',
);
const onboardingIntegrationCopy = document.querySelector<HTMLElement>(
  '#onboarding-integration-copy',
);
const onboardingJellyfinLANSetting = document.querySelector<HTMLElement>(
  '#onboarding-jellyfin-lan-setting',
);
const onboardingJellyfinLANCheckbox = document.querySelector<HTMLInputElement>(
  '#onboarding-jellyfin-lan',
);
const onboardingInstallButton = document.querySelector<HTMLButtonElement>('#onboarding-install');
const onboardingQualityList = document.querySelector<HTMLElement>('#onboarding-quality-list');
const onboardingQualitySummary = document.querySelector<HTMLElement>('#onboarding-quality-summary');
const onboardingQualityResult = document.querySelector<HTMLElement>('#onboarding-quality-result');
const onboardingQualityInstallButton = document.querySelector<HTMLButtonElement>(
  '#onboarding-quality-install',
);
const onboardingApplicationMessage = document.querySelector<HTMLElement>(
  '#onboarding-application-message',
);
const onboardingInstallationResult = document.querySelector<HTMLElement>(
  '#onboarding-installation-result',
);
const onboardingInstallationElement = document.querySelector<HTMLElement>(
  '#onboarding-installation',
);
const onboardingInstallationTitle = document.querySelector<HTMLElement>(
  '#onboarding-installation-title',
);
const onboardingInstallationProgressPanel = document.querySelector<HTMLElement>(
  '#onboarding-installation-progress',
);
const onboardingInstallationProgressSummary = document.querySelector<HTMLElement>(
  '#onboarding-installation-progress-summary',
);
const onboardingInstallationProgressTrack = document.querySelector<HTMLElement>(
  '#onboarding-installation-progress-track',
);
const onboardingInstallationProgressBar = document.querySelector<HTMLElement>(
  '#onboarding-installation-progress-bar',
);
const onboardingInstallationProgressList = document.querySelector<HTMLOListElement>(
  '#onboarding-installation-progress-list',
);
const onboardingInstallationCompletion = document.querySelector<HTMLElement>(
  '#onboarding-installation-completion',
);
const onboardingInstallationCompletionTitle = document.querySelector<HTMLElement>(
  '#onboarding-installation-completion-title',
);
const onboardingInstallationCompletionStatus = document.querySelector<HTMLElement>(
  '#onboarding-installation-completion-status',
);
const onboardingInstallationRetryButton = document.querySelector<HTMLButtonElement>(
  '#onboarding-installation-retry',
);
const onboardingOperationDetails = document.querySelector<HTMLDetailsElement>(
  '#onboarding-operation-details',
);
const onboardingOperationTechnical = document.querySelector<HTMLElement>(
  '#onboarding-operation-technical',
);
const onboardingCopySupportReport = document.querySelector<HTMLButtonElement>(
  '#onboarding-copy-support-report',
);
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
const infoView = document.querySelector<HTMLElement>('#info-view');
const showHomeButton = document.querySelector<HTMLButtonElement>('#show-home');
const showInfoButton = document.querySelector<HTMLButtonElement>('#show-info');
const showLicensesButton = document.querySelector<HTMLButtonElement>('#show-licenses');
const exportDiagnosticsButton = document.querySelector<HTMLButtonElement>('#export-diagnostics');
const licensesBackButton = document.querySelector<HTMLButtonElement>('#licenses-back');
const infoBackButton = document.querySelector<HTMLButtonElement>('#info-back');
const legalNoticesElement = document.querySelector<HTMLElement>('#legal-notices');
const infoCorsarrVersion = document.querySelector<HTMLElement>('#info-corsarr-version');
const infoQualityPolicy = document.querySelector<HTMLElement>('#info-quality-policy');
const infoRecyclarrVersion = document.querySelector<HTMLElement>('#info-recyclarr-version');
const infoTrashGuidesCommit = document.querySelector<HTMLElement>('#info-trash-guides-commit');
const infoOpenRecyclarr = document.querySelector<HTMLButtonElement>('#info-open-recyclarr');
const infoOpenTrashGuides = document.querySelector<HTMLButtonElement>('#info-open-trash-guides');
const infoAutomaticUpdates = document.querySelector<HTMLElement>('#info-automatic-updates');
const infoMessage = document.querySelector<HTMLElement>('#info-message');
const languageSelects = document.querySelectorAll<HTMLSelectElement>('.language-select');

for (const select of languageSelects) {
  select.value = currentLocale();
  select.addEventListener('change', async () => {
    const locale = normalizeLocale(select.value);
    if (!locale || locale === currentLocale()) return;
    for (const control of languageSelects) control.disabled = true;
    try {
      await SetLanguagePreference(locale);
      window.localStorage.setItem(languageStorageKey, locale);
      window.location.reload();
    } catch {
      for (const control of languageSelects) {
        control.value = currentLocale();
        control.disabled = false;
      }
    }
  });
}

let setupStatus: application.SetupStatus | undefined;
let availableApplications: Application[] = [];
let applicationCatalogState: 'loading' | 'ready' | 'error' = 'loading';
let selectedApplicationIDs = new Set<string>();
let selectionSaving = false;
let dashboardInstallingApplicationID: string | undefined;
let managedStatuses = new Map<string, ManagedStatus>();
let dataStatuses = new Map<string, DataStatus>();
let arrAccesses = new Map<string, application.ServiceAccessStatus>();
let qbittorrentAccess: application.ServiceAccessStatus | undefined;
let jellyfinAccess: application.ServiceAccessStatus | undefined;
let lazyLibrarianAccess: application.ServiceAccessStatus | undefined;
let jellyfinNetwork: main.JellyfinNetworkStatus | undefined;
let legalNotices: LegalNotice[] = [];
let qualityPresets: QualityPreset[] = [];
let currentRuntimeState = 'checking';
let currentHostReady = true;
let onboardingInstallationProgress: InstallationProgressItem[] = [];
let onboardingInstallationCompletionStage: 'waiting' | 'active' | 'ready' | 'failed' = 'waiting';

type OnboardingStep =
  | 'splash'
  | 'permissions'
  | 'environment'
  | 'storage'
  | 'applications'
  | 'quality'
  | 'installation';

const onboardingStepNumbers: Partial<Record<OnboardingStep, number>> = {
  permissions: 1,
  environment: 2,
  storage: 3,
  applications: 4,
  quality: 5,
};

const onboardingStepOrder: Record<OnboardingStep, number> = {
  splash: 0,
  permissions: 1,
  environment: 2,
  storage: 3,
  applications: 4,
  quality: 5,
  installation: 6,
};

function persistedOnboardingStep(): OnboardingStep {
  const step = setupStatus?.onboardingStep;
  if (!step || step === 'welcome' || step === 'complete') return 'splash';
  return step as OnboardingStep;
}

function onboardingHasAdvancedPast(step: OnboardingStep): boolean {
  return onboardingStepOrder[persistedOnboardingStep()] > onboardingStepOrder[step];
}

function hideOnboardingNotification(): void {
  if (onboardingNotification) onboardingNotification.hidden = true;
}

function showOnboardingNotification(message: string, tone: 'error' | 'success' = 'error'): void {
  if (!onboardingNotification || !onboardingNotificationMessage) return;
  onboardingNotificationMessage.textContent = message;
  onboardingNotification.classList.toggle('success', tone === 'success');
  if (onboardingNotificationIcon) {
    onboardingNotificationIcon.textContent = tone === 'success' ? '✓' : '!';
  }
  onboardingNotification.hidden = false;
}

const localizedIssueCodes = new Set([
  'application_configuration_failed',
  'application_install_failed',
  'application_status_unavailable',
  'application_update_failed',
  'application_update_rolled_back',
  'installation_failed',
  'onboarding_completion_failed',
  'post_quality_configuration_failed',
  'quality_profile_sync_failed',
  'runtime_storage_access_denied',
]);

function localizedIssue(issue: application.OperationIssue): application.OperationIssue {
  if (!localizedIssueCodes.has(issue.code)) return issue;
  return {
    code: issue.code,
    summary: t(`issue.${issue.code}.summary` as TranslationKey),
    nextAction: t(`issue.${issue.code}.next` as TranslationKey),
  };
}

function updateOnboardingCatalogAuthority(): void {
  if (onboardingRecommended) {
    onboardingRecommended.disabled = applicationCatalogState !== 'ready' || selectionSaving;
  }
  if (onboardingInstallButton) {
    onboardingInstallButton.disabled =
      applicationCatalogState !== 'ready' || !setupStatus?.canInstall;
    onboardingInstallButton.textContent = setupStatus?.qualityProfileRequired
      ? 'Continuar para qualidade'
      : 'Instalar aplicativos';
  }
  if (onboardingQualityInstallButton) {
    onboardingQualityInstallButton.disabled =
      !setupStatus?.canInstall || !setupStatus.qualityProfilePreset;
    onboardingQualityInstallButton.textContent = shouldManageQualityProfile(
      setupStatus?.qualityProfilePreset ?? '',
    )
      ? 'Instalar e aplicar perfil'
      : 'Instalar sem gerenciar perfis';
  }
}

function showOnboardingStep(step: OnboardingStep): void {
  for (const element of document.querySelectorAll<HTMLElement>('.onboarding-step')) {
    const active = element.id === `onboarding-${step}`;
    element.hidden = !active;
    element.inert = !active;
    element.setAttribute('aria-hidden', String(!active));
  }
  const stepNumber = onboardingStepNumbers[step];
  const totalSteps = onboardingStepTotal(setupStatus?.qualityProfileRequired ?? false);
  if (onboardingProgress) onboardingProgress.hidden = stepNumber === undefined;
  if (onboardingProgressLabel && stepNumber !== undefined) {
    onboardingProgressLabel.textContent = `Etapa ${stepNumber} de ${totalSteps}`;
  }
  if (onboardingProgress && stepNumber !== undefined) {
    const indicators = onboardingProgress.querySelectorAll<HTMLSpanElement>('span');
    for (const [index, indicator] of [...indicators].entries()) {
      indicator.hidden = index >= totalSteps;
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
  updateOnboardingCatalogAuthority();
  if (onboardingJellyfinLANSetting) {
    onboardingJellyfinLANSetting.hidden = !status.applications.includes('jellyfin');
  }
  if (onboardingJellyfinLANCheckbox) {
    onboardingJellyfinLANCheckbox.checked = status.jellyfinLanEnabled;
  }
  renderQualityProfileSelection();
}

function renderQualityProfileSelection(): void {
  if (!onboardingQualityList) return;
  onboardingQualityList.replaceChildren(
    ...qualityPresets.map((preset) => {
      const selected = preset.id === setupStatus?.qualityProfilePreset;
      const button = document.createElement('button');
      button.type = 'button';
      button.className = `quality-profile-card${selected ? ' selected' : ''}`;
      button.setAttribute('role', 'radio');
      button.setAttribute('aria-checked', String(selected));
      button.disabled = selectionSaving;
      const name = document.createElement('strong');
      name.textContent = preset.name;
      const description = document.createElement('span');
      description.textContent = preset.description;
      button.append(name, description);
      button.addEventListener('click', async () => {
        if (selectionSaving) return;
        selectionSaving = true;
        renderQualityProfileSelection();
        try {
          applySetupStatus(await SaveQualityProfilePreset(preset.id));
          if (onboardingQualityResult) {
            onboardingQualityResult.textContent = `${preset.name} selecionado.`;
            onboardingQualityResult.classList.remove('error');
          }
        } catch {
          if (onboardingQualityResult) {
            onboardingQualityResult.textContent = 'Não foi possível salvar o perfil escolhido.';
            onboardingQualityResult.classList.add('error');
          }
        } finally {
          selectionSaving = false;
          renderQualityProfileSelection();
          updateOnboardingCatalogAuthority();
        }
      });
      return button;
    }),
  );
  const selected = qualityPresets.find(({ id }) => id === setupStatus?.qualityProfilePreset);
  if (onboardingQualitySummary && selected) {
    onboardingQualitySummary.textContent = shouldManageQualityProfile(selected.id)
      ? selected.summary
      : `${selected.description} ${selected.summary}`;
  }
}

async function loadQualityProfilePresets(): Promise<void> {
  try {
    qualityPresets = await ListQualityProfilePresets();
  } catch {
    qualityPresets = [];
  }
  renderQualityProfileSelection();
  updateOnboardingCatalogAuthority();
}

function renderOnboardingIssue(issue?: application.OperationIssue): void {
  if (!onboardingOperationDetails || !onboardingOperationTechnical) return;
  onboardingOperationDetails.hidden = !issue;
  onboardingOperationDetails.open = false;
  if (onboardingCopySupportReport) onboardingCopySupportReport.hidden = !issue;
  if (onboardingCopySupportReport && issue) {
    onboardingCopySupportReport.textContent = 'Copiar log de erros';
  }
  const translated = issue ? localizedIssue(issue) : undefined;
  onboardingOperationTechnical.textContent = translated
    ? `${translated.summary}\n${translated.nextAction}\n${t('app.issueCode', { code: translated.code })}`
    : '';
}

onboardingCopySupportReport?.addEventListener('click', async () => {
  if (!onboardingCopySupportReport) return;
  onboardingCopySupportReport.disabled = true;
  try {
    await CopyLastInstallationSupportReport();
    onboardingCopySupportReport.textContent = 'Log copiado';
    showOnboardingNotification('Log de erros copiado', 'success');
  } catch {
    onboardingCopySupportReport.textContent = 'Não foi possível copiar';
    showOnboardingNotification(
      'Não foi possível copiar o log de erros. Tente novamente antes de fechar o Corsarr.',
    );
  } finally {
    onboardingCopySupportReport.disabled = false;
  }
});

function renderOperationIssue(issue?: application.OperationIssue): void {
  if (!operationDetailsElement || !operationTechnicalElement) return;
  operationDetailsElement.hidden = !issue;
  operationDetailsElement.open = false;
  const translated = issue ? localizedIssue(issue) : undefined;
  operationTechnicalElement.textContent = translated
    ? `${translated.summary}\n${translated.nextAction}\n${t('app.issueCode', { code: translated.code })}`
    : '';
}

function showView(view: 'home' | 'info' | 'licenses'): void {
  const showHome = view === 'home';
  const showInfo = view === 'info';
  if (homeView) homeView.hidden = !showHome;
  if (infoView) infoView.hidden = !showInfo;
  if (licensesView) licensesView.hidden = view !== 'licenses';
  showHomeButton?.classList.toggle('active', showHome);
  showInfoButton?.classList.toggle('active', showInfo);
  showLicensesButton?.classList.toggle('active', view === 'licenses');
  document.querySelector('.content')?.scrollTo({ top: 0, behavior: 'smooth' });
}

showHomeButton?.addEventListener('click', () => showView('home'));
showInfoButton?.addEventListener('click', () => showView('info'));
showLicensesButton?.addEventListener('click', () => showView('licenses'));
licensesBackButton?.addEventListener('click', () => showView('home'));
infoBackButton?.addEventListener('click', () => showView('home'));

async function openInfoComponentWebsite(
  button: HTMLButtonElement,
  componentID: string,
): Promise<void> {
  button.disabled = true;
  try {
    await OpenLegalLink(componentID, 'official');
  } catch {
    if (infoMessage) {
      infoMessage.textContent = t('site.openError');
      infoMessage.classList.add('error');
    }
  } finally {
    button.disabled = false;
  }
}

infoOpenRecyclarr?.addEventListener('click', () => {
  void openInfoComponentWebsite(infoOpenRecyclarr, 'runtime-recyclarr');
});
infoOpenTrashGuides?.addEventListener('click', () => {
  void openInfoComponentWebsite(infoOpenTrashGuides, 'guide-trash');
});

async function loadProductInfo(): Promise<void> {
  try {
    const productInfo = await GetProductInfo();
    if (infoCorsarrVersion) infoCorsarrVersion.textContent = productInfo.corsarrVersion;
    if (infoQualityPolicy) infoQualityPolicy.textContent = productInfo.qualityPolicyVersion;
    if (infoRecyclarrVersion) infoRecyclarrVersion.textContent = productInfo.recyclarrVersion;
    if (infoTrashGuidesCommit) infoTrashGuidesCommit.textContent = productInfo.trashGuidesCommit;
    if (infoAutomaticUpdates) {
      infoAutomaticUpdates.textContent = productInfo.automaticUpdates
        ? t('info.enabled')
        : t('info.disabled');
    }
    if (infoMessage) infoMessage.textContent = '';
  } catch {
    if (infoMessage) {
      infoMessage.textContent = t('info.loadError');
      infoMessage.classList.add('error');
    }
  }
}

exportDiagnosticsButton?.addEventListener('click', async () => {
  if (!exportDiagnosticsButton) return;
  exportDiagnosticsButton.disabled = true;
  try {
    const result = await ExportDiagnostics();
    if (!result.exported) return;
    if (messageElement) {
      messageElement.textContent = t('diagnostics.saved', { path: result.path });
      messageElement.classList.remove('error');
    }
    showView('home');
  } catch {
    if (messageElement) {
      messageElement.textContent = t('diagnostics.exportError');
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

const applicationIconURLs: Record<string, string> = {
  bazarr: new URL('./assets/application-icons/bazarr.svg', import.meta.url).href,
  fileflows: new URL('./assets/application-icons/fileflows.svg', import.meta.url).href,
  jellyfin: new URL('./assets/application-icons/jellyfin.svg', import.meta.url).href,
  jellyseerr: new URL('./assets/application-icons/seerr.svg', import.meta.url).href,
  lazylibrarian: new URL('./assets/application-icons/lazylibrarian.svg', import.meta.url).href,
  lidarr: new URL('./assets/application-icons/lidarr.svg', import.meta.url).href,
  prowlarr: new URL('./assets/application-icons/prowlarr.svg', import.meta.url).href,
  qbittorrent: new URL('./assets/application-icons/qbittorrent.svg', import.meta.url).href,
  radarr: new URL('./assets/application-icons/radarr.svg', import.meta.url).href,
  sonarr: new URL('./assets/application-icons/sonarr.svg', import.meta.url).href,
};

function createApplicationIcon(application: Application): HTMLElement {
  const icon = document.createElement('div');
  icon.className = `application-icon ${application.id}`;
  icon.setAttribute('aria-hidden', 'true');

  const fallback = document.createElement('span');
  fallback.className = 'application-icon-fallback';
  fallback.textContent = symbols[application.id] ?? application.name.slice(0, 2);
  icon.append(fallback);

  const iconURL = applicationIconURLs[application.id];
  if (!iconURL) return icon;

  const image = document.createElement('img');
  image.className = 'application-logo';
  image.alt = '';
  image.addEventListener('load', () => {
    fallback.hidden = true;
  });
  image.addEventListener('error', () => {
    image.remove();
  });
  image.src = iconURL;
  icon.append(image);
  return icon;
}

function createApplicationCard(application: Application): HTMLElement {
  const card = document.createElement('article');
  card.className = 'application-card';
  const managedStatus = managedStatuses.get(application.id);
  const installed = managedStatus?.state === 'running' || managedStatus?.state === 'stopped';
  const uncertainSelection =
    managedStatus?.state === 'attention' && selectedApplicationIDs.has(application.id);
  const selected = selectedApplicationIDs.has(application.id) || installed;
  if (selected) card.classList.add('selected');

  const icon = createApplicationIcon(application);

  const information = document.createElement('div');
  information.className = 'application-info';

  const title = document.createElement('h3');
  title.textContent = application.name;

  const description = document.createElement('p');
  description.textContent = application.description;

  information.append(title, description);

  if (managedStatus?.issue) {
    const issue = localizedIssue(managedStatus.issue);
    const details = document.createElement('details');
    details.className = 'application-status-details';
    const summary = document.createElement('summary');
    summary.textContent = t('app.details');
    const diagnostic = document.createElement('code');
    diagnostic.textContent = `${issue.summary}\n${issue.nextAction}\n${t('app.issueCode', { code: issue.code })}`;
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
    networkAddress.textContent = t('app.networkAddress', { url: jellyfinNetwork.urls[0] });
    networkAddress.title = jellyfinNetwork.urls.join('\n');
    information.append(networkAddress);
  }

  const actions = document.createElement('div');
  actions.className = 'application-actions';

  const selectButton = document.createElement('button');
  selectButton.className = 'select-button';
  selectButton.type = 'button';
  selectButton.textContent = installed
    ? t('app.installed')
    : !application.automatedSetup
      ? t('app.manualSetup')
      : dashboardInstallingApplicationID === application.id
        ? t('app.installing')
        : t('app.install');
  selectButton.setAttribute('aria-label', t('app.installAria', { name: application.name }));
  selectButton.disabled =
    selectionSaving || installed || uncertainSelection || !application.automatedSetup;
  if (installed) {
    selectButton.classList.add('installed-status');
    selectButton.title = t('app.installedTitle');
    selectButton.setAttribute('aria-label', t('app.installedAria', { name: application.name }));
  } else if (uncertainSelection) {
    selectButton.title = t('app.environmentBlocked');
    selectButton.setAttribute(
      'aria-label',
      t('app.environmentBlockedAria', { name: application.name }),
    );
  } else if (!application.automatedSetup) {
    selectButton.title = t('app.manualTitle');
    selectButton.setAttribute('aria-label', t('app.manualAria', { name: application.name }));
  }
  selectButton.addEventListener('click', async () => {
    if (selectionSaving) return;
    const previousSelection = new Set(selectedApplicationIDs);
    selectedApplicationIDs = new Set(
      selectApplicationWithIntegrations(
        selectedApplicationIDs,
        application.id,
        availableApplications,
      ),
    );
    selectionSaving = true;
    dashboardInstallingApplicationID = application.id;
    let selectionSaved = false;
    if (messageElement) {
      messageElement.textContent = t('app.preparingInstall', { name: application.name });
      messageElement.classList.remove('error');
    }
    renderApplications();

    try {
      const status = await SaveApplicationSelection([...selectedApplicationIDs]);
      selectionSaved = true;
      applySetupStatus(status);
      const result = await InstallSelectedApplications();
      await Promise.allSettled([
        loadApplicationStatuses(),
        loadApplicationDataStatuses(),
        loadJellyfinAccess(),
        loadLazyLibrarianAccess(),
        loadJellyfinNetwork(),
        loadQBittorrentAccess(),
        loadARRAccesses(),
      ]);
      if (!result.complete) {
        const failed = result.items.find((item) => item.failed);
        const issue = failed?.issue ? localizedIssue(failed.issue) : undefined;
        if (messageElement) {
          messageElement.textContent = `${issue?.summary ?? t('app.installIncomplete', { name: application.name })} ${issue?.nextAction ?? t('app.tryAgain')}`;
          messageElement.classList.add('error');
        }
        return;
      }
      if (messageElement) {
        messageElement.textContent = t('app.installReady', { name: application.name });
        messageElement.classList.remove('error');
      }
    } catch {
      if (!selectionSaved) selectedApplicationIDs = previousSelection;
      if (messageElement) {
        messageElement.textContent = t('app.installError', { name: application.name });
        messageElement.classList.add('error');
      }
    } finally {
      selectionSaving = false;
      dashboardInstallingApplicationID = undefined;
      renderApplications();
    }
  });

  const openButton = document.createElement('button');
  openButton.className = 'open-button';
  openButton.type = 'button';
  openButton.textContent = t('app.open');
  openButton.disabled = managedStatus?.state !== 'running';
  openButton.setAttribute('aria-label', t('app.openAria', { name: application.name }));
  openButton.addEventListener('click', async () => {
    openButton.disabled = true;
    messageElement?.classList.remove('error');
    if (messageElement) messageElement.textContent = t('app.opening', { name: application.name });

    try {
      await OpenApplication(application.id);
      if (messageElement) messageElement.textContent = t('app.opened', { name: application.name });
    } catch {
      if (messageElement) {
        messageElement.textContent = t('app.openError', { name: application.name });
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
      lifecycleButton(t('app.restart'), () => RestartApplication(application.id)),
      lifecycleButton(t('app.stop'), () => StopApplication(application.id)),
      applicationRemovalButton(application, managedStatus),
    );
  } else if (managedStatus?.state === 'stopped') {
    actions.append(
      lifecycleButton(t('app.start'), () => StartApplication(application.id)),
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
    arrAccesses.get(application.id)?.available &&
    (managedStatus?.state === 'running' || managedStatus?.state === 'stopped')
  ) {
    actions.append(arrCredentialButton(application));
  }
  if (
    application.id === 'jellyfin' &&
    jellyfinAccess?.available &&
    (managedStatus?.state === 'running' || managedStatus?.state === 'stopped')
  ) {
    actions.append(jellyfinCredentialButton());
  }
  if (
    application.id === 'jellyseerr' &&
    jellyfinAccess?.available &&
    (managedStatus?.state === 'running' || managedStatus?.state === 'stopped')
  ) {
    actions.append(seerrCredentialButton());
  }
  if (
    application.id === 'lazylibrarian' &&
    lazyLibrarianAccess?.available &&
    (managedStatus?.state === 'running' || managedStatus?.state === 'stopped')
  ) {
    actions.append(lazyLibrarianCredentialButton());
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
  button.textContent = t('app.update');
  button.addEventListener('click', async () => {
    const confirmed = window.confirm(t('app.updateConfirm', { name: target.name }));
    if (!confirmed) return;

    button.disabled = true;
    button.textContent = t('app.updating');
    if (messageElement) {
      messageElement.textContent = t('app.updatePreparing', { name: target.name });
      messageElement.classList.remove('error');
    }
    try {
      const result = await UpdateApplication(target.id);
      renderOperationIssue(result.issue);
      if (messageElement) {
        if (result.updated && !result.requiresAttention) {
          messageElement.textContent = t('app.updateReady', { name: target.name });
          messageElement.classList.remove('error');
        } else if (result.rolledBack) {
          messageElement.textContent = t('app.updateRolledBack', { name: target.name });
          messageElement.classList.add('error');
        } else if (result.requiresAttention) {
          messageElement.textContent = t('app.updateAttention', { name: target.name });
          messageElement.classList.add('error');
        } else {
          messageElement.textContent = t('app.updateCurrent', { name: target.name });
          messageElement.classList.remove('error');
        }
      }
      await loadApplicationStatuses();
    } catch {
      renderOperationIssue();
      if (messageElement) {
        messageElement.textContent = t('app.updateError', { name: target.name });
        messageElement.classList.add('error');
      }
    } finally {
      button.disabled = false;
      button.textContent = t('app.update');
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
        messageElement.textContent = t('app.lifecycleError', {
          action: label.toLocaleLowerCase(currentLocale()),
        });
        messageElement.classList.add('error');
      }
    } finally {
      button.disabled = false;
    }
  });
  return button;
}

async function removeApplication(target: Application): Promise<void> {
  const confirmed = window.confirm(t('app.removeConfirm', { name: target.name }));
  if (!confirmed) return;
  await RemoveApplication(target.id);
}

function applicationRemovalButton(target: Application, status: ManagedStatus): HTMLButtonElement {
  const blockers = status.removalBlockedBy ?? [];
  if (blockers.length === 0) {
    const button = lifecycleButton(t('app.remove'), () => removeApplication(target));
    button.classList.add('danger-button');
    return button;
  }

  const blockerNames = blockers.map(
    (id) => availableApplications.find((application) => application.id === id)?.name ?? id,
  );
  const button = document.createElement('button');
  button.className = 'lifecycle-button danger-button';
  button.type = 'button';
  button.textContent = t('app.remove');
  button.title = t('app.removeFirst', { names: blockerNames.join(', ') });
  button.setAttribute(
    'aria-label',
    t('app.removeBlockedAria', { name: target.name, reason: button.title }),
  );
  button.addEventListener('click', () => {
    window.alert(t('app.removeBlocked', { name: target.name, names: blockerNames.join(', ') }));
  });
  return button;
}

function dataRemovalButton(target: Application): HTMLButtonElement {
  const button = document.createElement('button');
  button.className = 'data-removal-button';
  button.type = 'button';
  button.textContent = t('app.removeData');
  button.addEventListener('click', async () => {
    const dataStatus = dataStatuses.get(target.id);
    const approximateSize = formatApproximateBytes(dataStatus?.sizeBytes ?? 0);
    const confirmed = window.confirm(
      t('app.removeDataConfirm', { size: approximateSize, name: target.name }),
    );
    if (!confirmed) return;

    button.disabled = true;
    try {
      const archived = await ArchiveApplicationData(target.id);
      if (messageElement) {
        messageElement.textContent = archived.archived
          ? t('app.dataArchived', { name: target.name })
          : t('app.noData', { name: target.name });
        messageElement.classList.remove('error');
      }
      await loadApplicationDataStatuses();
    } catch {
      if (messageElement) {
        messageElement.textContent = t('app.removeDataError', { name: target.name });
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
  return `${new Intl.NumberFormat(currentLocale(), { maximumFractionDigits }).format(value)} ${units[unitIndex]}`;
}

function credentialAccessControl(
  username: string,
  applicationName: string,
  copyPassword: () => Promise<void>,
  options: { usernameLabel?: string; buttonLabel?: string } = {},
): HTMLElement {
  const access = document.createElement('div');
  access.className = 'credential-access';
  const usernameLabel = document.createElement('span');
  usernameLabel.textContent = t('app.username');
  if (options.usernameLabel) {
    usernameLabel.textContent = options.usernameLabel;
  }
  const usernameValue = document.createElement('strong');
  usernameValue.textContent = username;
  const button = document.createElement('button');
  button.className = 'credential-button';
  button.type = 'button';
  const buttonLabel = options.buttonLabel ?? t('app.copyPassword');
  button.textContent = buttonLabel;
  button.addEventListener('click', async () => {
    button.disabled = true;
    try {
      await copyPassword();
      button.textContent = t('app.passwordCopied');
      if (messageElement) {
        messageElement.textContent = '';
        messageElement.classList.remove('error');
      }
      window.setTimeout(() => {
        button.textContent = buttonLabel;
      }, 2500);
    } catch {
      if (messageElement) {
        messageElement.textContent = t('app.copyPasswordError', { name: applicationName });
        messageElement.classList.add('error');
      }
    } finally {
      button.disabled = false;
    }
  });
  access.append(usernameLabel, usernameValue, button);
  return access;
}

function qbittorrentCredentialButton(): HTMLElement {
  return credentialAccessControl(
    qbittorrentAccess?.username ?? 'corsarr',
    'qBittorrent',
    CopyQBittorrentPassword,
  );
}

function arrCredentialButton(target: Application): HTMLElement {
  return credentialAccessControl(
    arrAccesses.get(target.id)?.username ?? 'corsarr',
    target.name,
    () => CopyARRPassword(target.id),
  );
}

function jellyfinCredentialButton(): HTMLElement {
  return credentialAccessControl(
    jellyfinAccess?.username ?? 'corsarr',
    'Jellyfin',
    CopyJellyfinPassword,
  );
}

function seerrCredentialButton(): HTMLElement {
  return credentialAccessControl(
    jellyfinAccess?.username ?? 'corsarr',
    t('seerr.account'),
    CopyJellyfinPassword,
    { usernameLabel: t('seerr.login'), buttonLabel: t('seerr.copyPassword') },
  );
}

function lazyLibrarianCredentialButton(): HTMLElement {
  return credentialAccessControl(
    lazyLibrarianAccess?.username ?? 'corsarr',
    'LazyLibrarian',
    CopyLazyLibrarianPassword,
  );
}

function jellyfinNetworkButton(url: string): HTMLButtonElement {
  const button = document.createElement('button');
  button.className = 'network-button';
  button.type = 'button';
  button.textContent = t('app.copyAddress');
  button.addEventListener('click', async () => {
    button.disabled = true;
    try {
      await CopyJellyfinNetworkURL(url);
      if (messageElement) {
        messageElement.textContent = t('app.addressCopied');
        messageElement.classList.remove('error');
      }
    } catch {
      if (messageElement) {
        messageElement.textContent = t('app.addressCopyError');
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

  const icon = createApplicationIcon(target);

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
  if (!target.automatedSetup) {
    metadata.textContent += ' · Configuração automática ainda indisponível';
  }
  const dependencies = target.dependencies ?? [];
  if (dependencies.length > 0) {
    const dependencyNames = dependencies.map(
      (id) => availableApplications.find((candidate) => candidate.id === id)?.name ?? id,
    );
    metadata.textContent += ` · Recomenda ${dependencyNames.join(', ')} para automatizar a configuração`;
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
  selectButton.disabled = selectionSaving || (!target.automatedSetup && !selected);
  selectButton.addEventListener('click', async () => {
    if (selectionSaving) return;
    const previousSelection = new Set(selectedApplicationIDs);
    selectedApplicationIDs = new Set(
      toggleApplicationSelection(selectedApplicationIDs, target.id, availableApplications),
    );
    selectionSaving = true;
    renderApplications();
    try {
      applySetupStatus(await SaveApplicationSelection([...selectedApplicationIDs]));
    } catch {
      selectedApplicationIDs = previousSelection;
      if (onboardingApplicationMessage) {
        onboardingApplicationMessage.textContent = 'Não foi possível salvar a seleção.';
        onboardingApplicationMessage.classList.add('error');
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

function renderApplicationCatalogError(): void {
  const error = document.createElement('div');
  error.className = 'onboarding-catalog-error';
  const copy = document.createElement('p');
  copy.textContent = 'Não foi possível carregar os aplicativos disponíveis.';
  const retry = document.createElement('button');
  retry.type = 'button';
  retry.className = 'secondary-button';
  retry.textContent = 'Tentar novamente';
  retry.addEventListener('click', () => void loadApplications());
  error.append(copy, retry);
  onboardingApplicationList?.replaceChildren(error);
}

function renderApplications(): void {
  const dashboardApplications = sortApplicationsByInstallation(availableApplications, [
    ...managedStatuses.values(),
  ]);
  applicationsElement?.replaceChildren(...dashboardApplications.map(createApplicationCard));
  onboardingApplicationList?.replaceChildren(
    ...availableApplications.map(createOnboardingApplicationCard),
  );
  renderIntegrationAdvice();
  if (onboardingCatalogCount) {
    onboardingCatalogCount.textContent = t('catalog.available', {
      count: availableApplications.length,
    });
  }
}

function renderRunningServicesSummary(): void {
  if (!heroTitleElement) return;
  const summary = runningServicesSummary([...managedStatuses.values()]);
  heroTitleElement.replaceChildren(
    t('dashboard.heroTitle'),
    document.createElement('br'),
    Object.assign(document.createElement('em'), {
      textContent: t('dashboard.runningServices', summary),
    }),
  );
}

function renderIntegrationAdvice(): void {
  if (!onboardingIntegrationGuidance || !onboardingIntegrationTitle || !onboardingIntegrationCopy)
    return;
  const missing = missingSelectedIntegrations(selectedApplicationIDs, availableApplications);
  onboardingIntegrationGuidance.classList.toggle('warning', missing.length > 0);
  onboardingIntegrationTitle.hidden = missing.length > 0;
  if (missing.length === 0) {
    onboardingIntegrationCopy.textContent =
      'Quando o Corsarr instala todos os aplicativos recomendados, ele pode conectar buscas e downloads para você. Ao escolher um aplicativo, também marcaremos as integrações recomendadas. Você pode desmarcar qualquer item se já usa seu próprio serviço.';
    return;
  }

  const capabilityByIntegration: Record<string, string> = {
    jellyfin: 'bibliotecas do Jellyfin',
    prowlarr: 'buscas pelo Prowlarr',
    qbittorrent: 'downloads pelo qBittorrent',
    radarr: 'filmes pelo Radarr',
    sonarr: 'séries pelo Sonarr',
  };
  const consequences = missing.map(({ consumerID, integrationID }) => {
    const consumer = availableApplications.find(({ id }) => id === consumerID)?.name ?? consumerID;
    return `${capabilityByIntegration[integrationID] ?? integrationID} para ${consumer}`;
  });
  onboardingIntegrationCopy.textContent = `Você está deixando parte da automação do Corsarr: ${consequences.join(', ')} precisarão de configuração manual. Sua escolha será respeitada.`;
}

async function loadApplications(): Promise<void> {
  if (!applicationsElement && !onboardingApplicationList) return;

  applicationCatalogState = 'loading';
  if (onboardingCatalogCount) onboardingCatalogCount.textContent = 'Carregando…';
  updateOnboardingCatalogAuthority();

  try {
    availableApplications = await ListApplications();
    if (availableApplications.length === 0) throw new Error('empty application catalog');
    applicationCatalogState = 'ready';
    renderApplications();
    if (countElement) {
      countElement.textContent = t('catalog.available', { count: availableApplications.length });
    }
  } catch {
    applicationCatalogState = 'error';
    availableApplications = [];
    applicationsElement?.replaceChildren();
    if (countElement) countElement.textContent = t('common.unavailable');
    if (onboardingCatalogCount) onboardingCatalogCount.textContent = 'Catálogo indisponível';
    renderApplicationCatalogError();
    if (messageElement) {
      messageElement.textContent = t('catalog.loadError');
      messageElement.classList.add('error');
    }
  } finally {
    updateOnboardingCatalogAuthority();
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
      kind.textContent =
        notice.componentType === 'runtime'
          ? 'Infraestrutura'
          : notice.componentType === 'asset'
            ? 'Recurso visual'
            : 'Aplicativo';
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
    renderRunningServicesSummary();
    updateJellyfinLANControl();
    renderApplications();
    renderLegalNotices();
  } catch {
    managedStatuses = new Map();
    renderRunningServicesSummary();
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

async function loadARRAccesses(): Promise<void> {
  try {
    const statuses = await GetARRAccessStatuses();
    arrAccesses = new Map(statuses.map((status) => [status.applicationId, status]));
    renderApplications();
  } catch {
    arrAccesses = new Map();
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

async function loadLazyLibrarianAccess(): Promise<void> {
  try {
    lazyLibrarianAccess = await GetLazyLibrarianAccessStatus();
    renderApplications();
  } catch {
    lazyLibrarianAccess = undefined;
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
    if (storageTitleElement) storageTitleElement.textContent = t('storage.saved');
    if (storageDescriptionElement) {
      storageDescriptionElement.textContent = t('storage.savedDescription');
    }
    if (storagePathElement) storagePathElement.textContent = status.storagePath;
    if (storageFactsElement) storageFactsElement.textContent = '';
    if (storageBadgeElement) {
      storageBadgeElement.textContent = t('storage.savedBadge');
      storageBadgeElement.className = 'runtime-badge ready';
    }
    if (chooseStorageButton) chooseStorageButton.textContent = t('storage.changeFolder');
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
      installationSummaryElement.textContent = `${applicationLabel}. O Corsarr instalará somente esta seleção e conectará os serviços compatíveis escolhidos juntos.`;
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
    let status = await GetSetupStatus();
    const persistedLocale = normalizeLocale(status.language);
    if (persistedLocale && persistedLocale !== currentLocale()) {
      window.localStorage.setItem(languageStorageKey, persistedLocale);
      window.location.reload();
      return;
    }
    if (!persistedLocale) {
      status = await SetLanguagePreference(currentLocale());
    }
    applySetupStatus(status);
  } catch {
    if (installationSummaryElement) {
      installationSummaryElement.textContent = t('setup.loadError');
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
    title: t('environment.readyTitle'),
    description: t('environment.readyDescription'),
    badge: t('environment.readyBadge'),
  },
  unavailable: {
    title: t('environment.requiredTitle'),
    description: t('environment.requiredDescription'),
    badge: t('environment.requiredBadge'),
  },
  stopped: {
    title: t('environment.stoppedTitle'),
    description: t('environment.stoppedDescription'),
    badge: t('environment.stoppedBadge'),
  },
  error: {
    title: t('environment.attentionTitle'),
    description: t('environment.attentionDescription'),
    badge: t('environment.attentionBadge'),
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
        : t('environment.hostAttention');
    }
    if (environmentDescriptionElement) {
      environmentDescriptionElement.textContent = currentHostReady
        ? runtimeMessage.description
        : environment.host.issues.join(' ');
    }
    if (environmentBadgeElement) {
      environmentBadgeElement.textContent = currentHostReady
        ? runtimeMessage.badge
        : t('environment.requirements');
      environmentBadgeElement.className = `runtime-badge ${currentHostReady ? environment.runtime.state : 'error'}`;
    }
    if (environmentTechnicalElement) {
      const details = [
        t('environment.provider', { provider: environment.runtime.provider }),
        t('environment.state', { state: environment.runtime.state }),
      ];
      if (environment.runtime.version) {
        details.push(t('environment.version', { version: environment.runtime.version }));
      }
      if (environment.runtime.technicalDetail) {
        details.push(t('environment.diagnostic', { detail: environment.runtime.technicalDetail }));
      }
      if (environment.host.osVersion) details.push(`macOS: ${environment.host.osVersion}`);
      if (environment.host.memoryBytes) {
        details.push(
          t('environment.memory', {
            amount: (environment.host.memoryBytes / 1024 ** 3).toFixed(1),
          }),
        );
      }
      if (environment.host.freeDiskBytes) {
        details.push(
          t('environment.disk', {
            amount: (environment.host.freeDiskBytes / 1024 ** 3).toFixed(1),
          }),
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
        environment.runtime.state === 'stopped'
          ? t('environment.start')
          : t('dashboard.prepareComputer');
    }
    if (onboardingEnvironmentTitle) {
      onboardingEnvironmentTitle.textContent = currentHostReady
        ? runtimeMessage.title
        : t('environment.hostAttention');
    }
    if (onboardingEnvironmentDescription) {
      onboardingEnvironmentDescription.textContent = currentHostReady
        ? runtimeMessage.description
        : environment.host.issues.join(' ');
    }
    if (onboardingEnvironmentBadge) {
      onboardingEnvironmentBadge.textContent = currentHostReady
        ? runtimeMessage.badge
        : t('environment.requirements');
      onboardingEnvironmentBadge.className = `runtime-badge ${currentHostReady ? environment.runtime.state : 'error'}`;
    }
    if (onboardingPrepareRuntime) {
      onboardingPrepareRuntime.hidden = environment.runtime.state === 'ready';
      onboardingPrepareRuntime.disabled = !currentHostReady;
      onboardingPrepareRuntime.textContent =
        environment.runtime.state === 'stopped'
          ? t('environment.start')
          : t('dashboard.prepareComputer');
    }
    if (onboardingEnvironmentNext) {
      onboardingEnvironmentNext.disabled =
        !currentHostReady || environment.runtime.state !== 'ready';
    }
    updateInstallAuthority();
  } catch {
    currentRuntimeState = 'error';
    currentHostReady = false;
    if (environmentTitleElement)
      environmentTitleElement.textContent = t('environment.attentionTitle');
    if (environmentDescriptionElement) {
      environmentDescriptionElement.textContent = t('environment.retry');
    }
    if (environmentBadgeElement) {
      environmentBadgeElement.textContent = t('environment.attentionBadge');
      environmentBadgeElement.className = 'runtime-badge error';
    }
    if (onboardingEnvironmentTitle) {
      onboardingEnvironmentTitle.textContent = t('environment.attentionTitle');
    }
    if (onboardingEnvironmentDescription) {
      onboardingEnvironmentDescription.textContent = t('environment.retry');
    }
    if (onboardingEnvironmentBadge) {
      onboardingEnvironmentBadge.textContent = t('environment.attentionBadge');
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
      installationResultElement.textContent = t('runtime.resolveRequirements');
      installationResultElement.classList.add('error');
    }
    return;
  }
  if (!setupStatus?.termsAccepted) {
    if (messageElement) {
      messageElement.textContent = t('runtime.authorizationUnavailable');
      messageElement.classList.add('error');
    }
    return;
  }
  if (currentRuntimeState === 'unavailable' && !window.confirm(t('runtime.installConfirm'))) {
    return;
  }

  prepareRuntimeButton.disabled = true;
  prepareRuntimeButton.textContent =
    currentRuntimeState === 'unavailable' ? t('runtime.installing') : t('runtime.starting');
  if (environmentDescriptionElement) {
    environmentDescriptionElement.textContent =
      currentRuntimeState === 'unavailable'
        ? t('runtime.downloading')
        : t('runtime.startingDescription');
  }
  try {
    applySetupStatus(await AcceptCurrentTerms());
    const result = await PrepareRuntime();
    if (!result.ready) throw new Error('runtime not ready');
    await loadEnvironment();
    if (messageElement) {
      messageElement.textContent = result.installed
        ? t('runtime.installedReady')
        : t('runtime.startedReady');
      messageElement.classList.remove('error');
    }
  } catch {
    if (environmentDescriptionElement) {
      environmentDescriptionElement.textContent = t('runtime.incomplete');
    }
    if (messageElement) {
      messageElement.textContent = t('runtime.error');
      messageElement.classList.add('error');
    }
  } finally {
    prepareRuntimeButton.disabled = false;
    if (currentRuntimeState !== 'ready') {
      prepareRuntimeButton.textContent =
        currentRuntimeState === 'stopped' ? t('environment.start') : t('dashboard.prepareComputer');
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
        loadLazyLibrarianAccess(),
        loadJellyfinNetwork(),
        loadQBittorrentAccess(),
        loadARRAccesses(),
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

onboardingNotificationDismiss?.addEventListener('click', hideOnboardingNotification);

onboardingPermissionsNext?.addEventListener('click', async () => {
  if (!onboardingTermsCheckbox?.checked || !onboardingPermissionsNext) return;
  const resumeExistingProgress = onboardingHasAdvancedPast('permissions');
  const startAtLogin = onboardingStartLoginCheckbox?.checked ?? false;
  onboardingPermissionsNext.disabled = true;
  hideOnboardingNotification();
  try {
    applySetupStatus(await AcceptCurrentTerms());
    if (setupStatus?.startAtLoginSupported) {
      applySetupStatus(await SetStartAtLogin(startAtLogin));
    }
    if (!resumeExistingProgress) applySetupStatus(await AdvanceOnboarding());
  } catch {
    try {
      applySetupStatus(await GetSetupStatus());
    } catch {
      // Preserve the last known safe state when even the refresh is unavailable.
    }
    showOnboardingStep('permissions');
    showOnboardingNotification(
      'Não foi possível salvar sua autorização. Revise as opções e tente continuar novamente. Nenhum componente foi instalado.',
    );
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
  if (!onboardingRecommended || selectionSaving || applicationCatalogState !== 'ready') return;
  onboardingRecommended.disabled = true;
  selectionSaving = true;
  renderApplications();
  try {
    applySetupStatus(await SelectRecommendedApplications());
    if (onboardingApplicationMessage) {
      onboardingApplicationMessage.textContent =
        'Configuração recomendada selecionada. Você ainda pode personalizá-la.';
      onboardingApplicationMessage.classList.remove('error');
    }
  } catch {
    if (onboardingApplicationMessage) {
      onboardingApplicationMessage.textContent =
        'Não foi possível selecionar a configuração recomendada.';
      onboardingApplicationMessage.classList.add('error');
    }
  } finally {
    selectionSaving = false;
    updateOnboardingCatalogAuthority();
    renderApplications();
  }
});

onboardingJellyfinLANCheckbox?.addEventListener('change', async () => {
  if (!onboardingJellyfinLANCheckbox) return;
  const enabled = onboardingJellyfinLANCheckbox.checked;
  onboardingJellyfinLANCheckbox.disabled = true;
  hideOnboardingNotification();
  try {
    applySetupStatus(await SetJellyfinLAN(enabled));
  } catch {
    try {
      applySetupStatus(await GetSetupStatus());
    } catch {
      // Keep the last known state when the setup refresh is also unavailable.
    }
    showOnboardingStep('applications');
    showOnboardingNotification(
      enabled
        ? 'Não foi possível liberar o Jellyfin para esta rede. Confirme que o Jellyfin está selecionado e tente novamente.'
        : 'Não foi possível remover o acesso do Jellyfin à rede. Tente novamente.',
    );
  } finally {
    onboardingJellyfinLANCheckbox.disabled = false;
  }
});

function installationStageLabel(stage: TrackedInstallationStage): string {
  const labels: Record<TrackedInstallationStage, string> = {
    waiting: 'Aguardando',
    installing: 'Baixando e iniciando',
    provisioning: 'Configurando',
    ready: 'Pronto',
    failed: 'Precisa de atenção',
  };
  return labels[stage];
}

function renderOnboardingInstallationProgress(): void {
  if (
    !onboardingInstallationProgressPanel ||
    !onboardingInstallationProgressSummary ||
    !onboardingInstallationProgressList
  ) {
    return;
  }

  const ready = onboardingInstallationProgress.filter(({ stage }) => stage === 'ready').length;
  const failed = onboardingInstallationProgress.some(({ stage }) => stage === 'failed');
  const active = onboardingInstallationProgress.find(
    ({ stage }) => stage === 'installing' || stage === 'provisioning',
  );
  const totalSteps = onboardingInstallationProgress.length + 1;
  const activeWeight = active ? 0.5 : onboardingInstallationCompletionStage === 'active' ? 0.5 : 0;
  const completedSteps = ready + (onboardingInstallationCompletionStage === 'ready' ? 1 : 0);
  const progress = Math.round(((completedSteps + activeWeight) / Math.max(1, totalSteps)) * 100);

  onboardingInstallationProgressSummary.textContent =
    failed || onboardingInstallationCompletionStage === 'failed'
      ? 'Precisa de atenção'
      : onboardingInstallationCompletionStage === 'ready'
        ? 'Concluído'
        : onboardingInstallationCompletionStage === 'active'
          ? 'Finalizando'
          : active
            ? `${active.position} de ${onboardingInstallationProgress.length} aplicativos`
            : 'Preparando';
  if (onboardingInstallationProgressBar) {
    onboardingInstallationProgressBar.style.width = `${Math.max(4, progress)}%`;
  }
  onboardingInstallationProgressTrack?.setAttribute('aria-valuenow', String(progress));

  onboardingInstallationProgressList.replaceChildren(
    ...onboardingInstallationProgress.map((item) => {
      const applicationName =
        availableApplications.find(({ id }) => id === item.applicationId)?.name ??
        item.applicationId;
      const row = document.createElement('li');
      row.className = `installation-progress-item ${item.stage}`;
      const indicator = document.createElement('span');
      indicator.className = 'installation-progress-indicator';
      indicator.setAttribute('aria-hidden', 'true');
      const copy = document.createElement('div');
      const name = document.createElement('strong');
      name.textContent = applicationName;
      const status = document.createElement('small');
      status.textContent = installationStageLabel(item.stage);
      copy.append(name, status);
      row.append(indicator, copy);
      return row;
    }),
  );

  const managesQuality =
    setupStatus?.qualityProfileRequired &&
    shouldManageQualityProfile(setupStatus.qualityProfilePreset ?? '');
  if (onboardingInstallationCompletionTitle) {
    onboardingInstallationCompletionTitle.textContent = managesQuality
      ? 'Perfil de qualidade'
      : 'Finalizando configuração';
  }
  if (onboardingInstallationCompletion) {
    onboardingInstallationCompletion.className = `installation-completion ${onboardingInstallationCompletionStage}`;
  }
  if (onboardingInstallationCompletionStatus) {
    const completionLabels = managesQuality
      ? {
          waiting: 'Aguardando os aplicativos',
          active: 'Aplicando o perfil selecionado',
          ready: 'Perfil aplicado',
          failed: 'Não foi possível concluir',
        }
      : {
          waiting: 'Aguardando os aplicativos',
          active: 'Concluindo a configuração inicial',
          ready: 'Configuração concluída',
          failed: 'Não foi possível concluir',
        };
    onboardingInstallationCompletionStatus.textContent =
      completionLabels[onboardingInstallationCompletionStage];
  }
}

function startOnboardingInstallationProgress(): void {
  const resumesFinalization =
    onboardingInstallationProgress.length > 0 &&
    onboardingInstallationProgress.every(({ stage }) => stage === 'ready');
  if (!resumesFinalization) {
    onboardingInstallationProgress = createInstallationProgress(selectedApplicationIDs);
  }
  onboardingInstallationCompletionStage = resumesFinalization ? 'active' : 'waiting';
  renderOnboardingInstallationProgress();
}

const genericOnboardingInstallationIssue = {
  code: 'installation_failed',
  summary: 'A instalação ou configuração não terminou.',
  nextAction: 'Copie o log de erros para entender o que aconteceu.',
} as application.OperationIssue;

function showOnboardingInstallationFailure(
  message: string,
  issue: application.OperationIssue = genericOnboardingInstallationIssue,
): void {
  renderOnboardingIssue(issue);
  onboardingInstallationCompletionStage = 'failed';
  renderOnboardingInstallationProgress();
  if (onboardingInstallationTitle) {
    onboardingInstallationTitle.textContent = 'A instalação precisa de atenção.';
  }
  if (onboardingInstallationResult) {
    onboardingInstallationResult.textContent = message;
    onboardingInstallationResult.classList.add('error');
  }
  if (onboardingInstallationRetryButton) {
    onboardingInstallationRetryButton.hidden = false;
    onboardingInstallationRetryButton.focus();
  }
  if (onboardingCopySupportReport) onboardingCopySupportReport.hidden = false;
}

async function installSelectedFromOnboarding(): Promise<void> {
  if (!setupStatus?.canInstall) return;
  showOnboardingStep('installation');
  if (onboardingInstallationRetryButton) {
    onboardingInstallationRetryButton.hidden = true;
    onboardingInstallationRetryButton.disabled = true;
  }
  renderOnboardingIssue();
  if (onboardingInstallationTitle) {
    onboardingInstallationTitle.textContent = 'Preparando seu servidor.';
    onboardingInstallationTitle.focus();
  }
  startOnboardingInstallationProgress();
  if (onboardingInstallationResult) {
    onboardingInstallationResult.textContent =
      onboardingInstallationCompletionStage === 'active'
        ? 'Aplicativos já estão prontos. Tentando novamente a configuração final…'
        : 'Verificando o ambiente e preparando a instalação. Isso pode levar alguns minutos.';
    onboardingInstallationResult.classList.remove('error');
  }
  try {
    const result = await InstallSelectedApplications();
    if (!result.complete) {
      const failed = result.items.find((item) => item.failed);
      if (failed?.issue?.code === 'runtime_storage_access_denied') {
        if (onboardingStorageTitle) onboardingStorageTitle.textContent = 'Escolha outra pasta';
        if (onboardingStorageDescription) {
          onboardingStorageDescription.textContent = failed.issue.summary;
        }
        if (onboardingStorageMessage) {
          onboardingStorageMessage.textContent = failed.issue.nextAction;
          onboardingStorageMessage.classList.add('error');
        }
        if (onboardingStorageBadge) {
          onboardingStorageBadge.textContent = 'Acesso necessário';
          onboardingStorageBadge.className = 'runtime-badge error';
        }
        showOnboardingStep('storage');
        showOnboardingNotification(
          'A pasta continua preservada, mas o Docker precisa conseguir acessá-la.',
        );
        if (onboardingInstallationRetryButton) {
          onboardingInstallationRetryButton.hidden = true;
        }
        return;
      }
      showOnboardingInstallationFailure(
        `${failed?.issue?.summary ?? 'A instalação não terminou.'} ${failed?.issue?.nextAction ?? 'Tente novamente.'}`,
        failed?.issue,
      );
      return;
    }
    onboardingInstallationCompletionStage = 'ready';
    renderOnboardingInstallationProgress();
    if (onboardingInstallationTitle) {
      onboardingInstallationTitle.textContent = 'Seu servidor está pronto.';
    }
    if (onboardingInstallationResult) {
      onboardingInstallationResult.textContent = 'Instalação e configuração concluídas.';
    }
    applySetupStatus(await GetSetupStatus());
    await Promise.all([
      loadApplicationStatuses(),
      loadApplicationDataStatuses(),
      loadJellyfinAccess(),
      loadLazyLibrarianAccess(),
      loadJellyfinNetwork(),
      loadQBittorrentAccess(),
      loadARRAccesses(),
    ]);
    if (messageElement) {
      messageElement.textContent = `${result.items.length} aplicativos instalados. A configuração inicial foi concluída.`;
      messageElement.classList.remove('error');
    }
  } catch {
    showOnboardingInstallationFailure(
      'Não foi possível concluir a instalação. Suas escolhas foram preservadas.',
    );
  } finally {
    if (onboardingInstallationRetryButton) {
      onboardingInstallationRetryButton.disabled = false;
    }
    if (!setupStatus?.onboardingCompleted) {
      updateOnboardingCatalogAuthority();
    }
  }
}

onboardingInstallButton?.addEventListener('click', async () => {
  if (!onboardingInstallButton || !setupStatus?.canInstall) return;
  if (setupStatus.qualityProfileRequired) {
    if (onboardingHasAdvancedPast('applications')) {
      syncOnboarding(setupStatus);
    } else {
      applySetupStatus(await AdvanceOnboarding());
    }
    return;
  }
  await installSelectedFromOnboarding();
});

onboardingQualityInstallButton?.addEventListener('click', async () => {
  if (!onboardingQualityInstallButton) return;
  await installSelectedFromOnboarding();
});

onboardingInstallationRetryButton?.addEventListener('click', async () => {
  await installSelectedFromOnboarding();
});

async function loadInitialState(): Promise<void> {
  await Promise.all([loadEnvironment(), loadSetup()]);
  await Promise.all([
    loadApplications(),
    loadLegalNotices(),
    loadQualityProfilePresets(),
    loadProductInfo(),
  ]);
  await Promise.all([
    loadApplicationStatuses(),
    loadApplicationDataStatuses(),
    loadJellyfinAccess(),
    loadLazyLibrarianAccess(),
    loadJellyfinNetwork(),
    loadQBittorrentAccess(),
    loadARRAccesses(),
  ]);
}

EventsOn('corsarr:background-recovery-complete', () => {
  void Promise.all([
    loadEnvironment(),
    loadApplicationStatuses(),
    loadJellyfinAccess(),
    loadLazyLibrarianAccess(),
    loadJellyfinNetwork(),
    loadQBittorrentAccess(),
    loadARRAccesses(),
  ]);
});

EventsOn('corsarr:installation-progress', (progress: InstallationProgressEvent) => {
  const applicationName =
    availableApplications.find((application) => application.id === progress.applicationId)?.name ??
    'aplicativo';
  const stageMessages: Record<InstallationProgressEvent['stage'], string> = {
    installing: `Baixando e iniciando ${applicationName}`,
    provisioning: `Configurando ${applicationName}`,
    ready: `${applicationName} está pronto`,
    failed: `${applicationName} precisa de atenção`,
  };
  if (installationResultElement) {
    installationResultElement.textContent = `${stageMessages[progress.stage]} (${progress.position} de ${progress.total}).`;
    installationResultElement.classList.toggle('error', progress.stage === 'failed');
  }
  if (dashboardInstallingApplicationID && messageElement) {
    messageElement.textContent = `${stageMessages[progress.stage]} (${progress.position} de ${progress.total}).`;
    messageElement.classList.toggle('error', progress.stage === 'failed');
  }
  if (onboardingInstallationResult && !onboardingInstallationElement?.hidden) {
    onboardingInstallationResult.textContent = `${stageMessages[progress.stage]} (${progress.position} de ${progress.total}).`;
    onboardingInstallationResult.classList.toggle('error', progress.stage === 'failed');
    onboardingInstallationProgress = applyInstallationProgress(
      onboardingInstallationProgress,
      progress,
    );
    if (progress.stage === 'ready' && progress.position === progress.total) {
      onboardingInstallationCompletionStage = 'active';
      onboardingInstallationResult.textContent =
        setupStatus?.qualityProfileRequired &&
        shouldManageQualityProfile(setupStatus.qualityProfilePreset ?? '')
          ? 'Aplicativos prontos. Aplicando o perfil de qualidade selecionado…'
          : 'Aplicativos prontos. Finalizando a configuração inicial…';
    }
    renderOnboardingInstallationProgress();
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
