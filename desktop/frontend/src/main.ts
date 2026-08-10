import './style.css';
import { ListApplications, OpenApplication } from '../wailsjs/go/main/App';
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
  '      <div class="machine"><span class="machine-icon" aria-hidden="true">⌘</span><span><small>Executando neste Mac</small><strong>Apple Silicon</strong></span></div>',
  '    </header>',
  '    <section class="hero" aria-labelledby="hero-title">',
  '      <div>',
  '        <span class="phase">FASE 1 · FUNDAÇÃO</span>',
  '        <h2 id="hero-title">Tudo no lugar certo,<br><em>sem complicação.</em></h2>',
  '        <p>O Corsarr está preparando uma experiência única para instalar e cuidar dos seus aplicativos de mídia.</p>',
  '      </div>',
  '      <div class="radar" aria-hidden="true"><span></span><span></span><i></i><b>C</b></div>',
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

void loadApplications();
