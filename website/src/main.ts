import "./brand.css";
import "./style.css";

type Locale = "pt-BR" | "en";

type Copy = {
  nav: { how: string; apps: string; safety: string; faq: string; github: string; language: string };
  hero: {
    eyebrow: string;
    title: string;
    accent: string;
    body: string;
    download: string;
    cli: string;
    note: string;
  };
  preview: {
    eyebrow: string;
    title: string;
    status: string;
    services: string;
    caption: string;
    onboardingCaption: string;
  };
  problem: { eyebrow: string; title: string; body: string; items: string[] };
  steps: { eyebrow: string; title: string; body: string; items: Array<{ title: string; body: string }> };
  apps: { eyebrow: string; title: string; body: string; trademark: string };
  safety: { eyebrow: string; title: string; body: string; cards: Array<{ title: string; body: string }> };
  responsible: { title: string; body: string };
  cli: { eyebrow: string; title: string; body: string; command: string; action: string };
  downloads: {
    eyebrow: string;
    title: string;
    body: string;
    platforms: Array<{ name: string; detail: string; action: string; href: string }>;
    notice: string;
  };
  faq: { eyebrow: string; title: string; items: Array<{ question: string; answer: string }> };
  footer: { body: string; github: string; licenses: string; copyright: string };
};

const github = "https://github.com/woliveiras/corsarr";
const releases = `${github}/releases`;

const copy: Record<Locale, Copy> = {
  "pt-BR": {
    nav: {
      how: "Como funciona",
      apps: "Aplicativos",
      safety: "Controle local",
      faq: "Dúvidas",
      github: "GitHub",
      language: "English",
    },
    hero: {
      eyebrow: "AUTOMAÇÃO DE MÍDIA NO SEU COMPUTADOR",
      title: "Seu servidor de mídia,",
      accent: "sem complicação.",
      body: "Configure uma experiência completa para organizar, baixar legendas e assistir aos seus filmes e séries — sem editar YAML, decorar portas ou administrar containers.",
      download: "Baixar Corsarr",
      cli: "Conhecer a versão CLI",
      note: "Desktop para macOS, Windows e Linux · CLI para usuários avançados",
    },
    preview: {
      eyebrow: "CONTROLE LOCAL",
      title: "Tudo funcionando,",
      status: "7/7 serviços rodando",
      services: "Aplicativos conectados e prontos",
      caption: "Interface real do Corsarr Desktop em execução no macOS.",
      onboardingCaption: "Configuração inicial real, depois de remover a instalação anterior com segurança.",
    },
    problem: {
      eyebrow: "UMA EXPERIÊNCIA, NÃO UMA PILHA DE CONFIGURAÇÕES",
      title: "Cuide da sua mídia. O Corsarr cuida da infraestrutura.",
      body: "Um servidor de mídia costuma exigir várias ferramentas, caminhos e credenciais. O Corsarr reúne essa preparação em uma jornada clara.",
      items: [
        "Sem Docker na experiência principal",
        "Sem arquivos YAML para editar",
        "Sem conectar cada aplicativo à mão",
      ],
    },
    steps: {
      eyebrow: "COMO FUNCIONA",
      title: "Do computador vazio à sua biblioteca organizada.",
      body: "Você escolhe o que quer usar. O Corsarr prepara o restante e mantém os aplicativos acessíveis em suas interfaces originais.",
      items: [
        {
          title: "Escolha onde guardar",
          body: "Aponte uma pasta. O Corsarr verifica espaço, escrita e eficiência antes de criar qualquer coisa.",
        },
        {
          title: "Monte sua experiência",
          body: "Escolha filmes, séries, legendas, pedidos e streaming — ou use a configuração recomendada.",
        },
        {
          title: "Instale em um fluxo",
          body: "O Corsarr instala os aplicativos, cria caminhos consistentes e conecta as integrações compatíveis.",
        },
        {
          title: "Assista e administre",
          body: "Abra, atualize, reinicie ou remova aplicativos sem perder sua biblioteca e suas configurações.",
        },
      ],
    },
    apps: {
      eyebrow: "APLICATIVOS QUE JÁ TRABALHAM JUNTOS",
      title: "As melhores ferramentas, conectadas para você.",
      body: "Corsarr não substitui esses projetos. Ele oferece um caminho simples e cuidadoso para instalá-los e conectá-los.",
      trademark:
        "Marcas e logos pertencem aos respectivos projetos. Corsarr não é afiliado nem endossado por eles.",
    },
    safety: {
      eyebrow: "LOCAL POR PADRÃO",
      title: "Seu computador. Seus arquivos. Suas escolhas.",
      body: "A simplicidade não precisa esconder decisões importantes. O Corsarr reduz a operação técnica enquanto mantém dados e consentimento sob seu controle.",
      cards: [
        {
          title: "Privado por padrão",
          body: "Painéis administrativos ficam disponíveis apenas neste computador. Acesso do Jellyfin à rede local é uma escolha separada.",
        },
        {
          title: "Dados preservados",
          body: "Remover um aplicativo não apaga sua mídia. Configuração e biblioteca têm ações distintas e explícitas.",
        },
        {
          title: "Componentes revisados",
          body: "O catálogo usa versões identificadas e contratos de execução validados antes de tocar nos serviços existentes.",
        },
      ],
    },
    responsible: {
      title: "Feito para a sua biblioteca legítima.",
      body: "Corsarr não inclui filmes, séries, indexadores ou contas de terceiros. Use-o somente com conteúdo e fontes que você tem direito de acessar, respeitando as leis e os termos aplicáveis.",
    },
    cli: {
      eyebrow: "PREFERE O TERMINAL?",
      title: "A mesma bandeira, com controle avançado.",
      body: "A versão CLI continua disponível para servidores, laboratórios domésticos e pessoas que desejam gerar e revisar sua configuração Docker Compose.",
      command: "corsarr generate",
      action: "Documentação da CLI",
    },
    downloads: {
      eyebrow: "ESCOLHA SUA PLATAFORMA",
      title: "Pronto para começar?",
      body: "As compilações Desktop são publicadas junto com as releases do Corsarr. Escolha a plataforma deste computador.",
      platforms: [
        {
          name: "macOS",
          detail: "Apple Silicon + Intel",
          action: "Baixar para macOS",
          href: `${releases}/latest/download/corsarr_desktop_darwin_universal.zip`,
        },
        {
          name: "Windows",
          detail: "Windows 10/11 · x64",
          action: "Baixar para Windows",
          href: `${releases}/latest/download/corsarr_desktop_windows_amd64.zip`,
        },
        {
          name: "Linux",
          detail: "Linux · x64",
          action: "Baixar para Linux",
          href: `${releases}/latest/download/corsarr_desktop_linux_amd64.tar.gz`,
        },
      ],
      notice: "As primeiras versões podem não ser assinadas. Confira as notas da release antes de instalar.",
    },
    faq: {
      eyebrow: "DÚVIDAS FREQUENTES",
      title: "Antes de içar as velas.",
      items: [
        {
          question: "Preciso saber usar Docker?",
          answer:
            "Não para usar o Corsarr Desktop. O runtime continua aparecendo em termos, licenças e detalhes técnicos, mas a jornada principal fala sobre aplicativos, armazenamento e mídia.",
        },
        {
          question: "O Corsarr fornece filmes ou séries?",
          answer:
            "Não. O Corsarr instala e conecta software. Ele não fornece mídia, indexadores, contas, VPNs ou fontes de download. Você é responsável por usar apenas conteúdo e serviços permitidos.",
        },
        {
          question: "Posso assistir na minha TV?",
          answer:
            "Sim. Quando o Jellyfin está selecionado, o Corsarr pode liberar apenas o streaming para aparelhos da mesma rede local. Os painéis administrativos continuam privados.",
        },
        {
          question: "O que acontece se eu remover um aplicativo?",
          answer:
            "O container é removido, mas configurações, downloads e biblioteca são preservados. A exclusão da configuração é uma ação separada e nunca inclui a mídia compartilhada.",
        },
        {
          question: "Windows e Linux são suportados?",
          answer:
            "As compilações estão previstas para as três plataformas. O onboarding completo começa pelo macOS; veja as notas de cada release para o estado atual de Windows e Linux.",
        },
        {
          question: "Ainda existe uma versão CLI?",
          answer:
            "Sim. A CLI permanece disponível para usuários avançados e pode gerar a configuração Docker Compose da stack.",
        },
      ],
    },
    footer: {
      body: "Navegue pelos altos mares da automação de mídia com simplicidade e controle local.",
      github: "Código no GitHub",
      licenses: "Licenças e créditos",
      copyright: "Projeto independente e open source.",
    },
  },
  en: {
    nav: {
      how: "How it works",
      apps: "Apps",
      safety: "Local control",
      faq: "FAQ",
      github: "GitHub",
      language: "Português",
    },
    hero: {
      eyebrow: "MEDIA AUTOMATION ON YOUR COMPUTER",
      title: "Your media server,",
      accent: "made simple.",
      body: "Set up a complete experience to organize, download subtitles, and watch your movies and shows — without editing YAML, memorizing ports, or managing containers.",
      download: "Download Corsarr",
      cli: "Explore the CLI",
      note: "Desktop for macOS, Windows, and Linux · CLI for advanced users",
    },
    preview: {
      eyebrow: "LOCAL CONTROL",
      title: "Everything is running,",
      status: "7/7 services online",
      services: "Apps connected and ready",
      caption: "The real Corsarr Desktop interface running on macOS.",
      onboardingCaption: "The real first-run experience after safely removing the previous installation.",
    },
    problem: {
      eyebrow: "ONE EXPERIENCE, NOT A STACK OF CONFIG FILES",
      title: "Care for your media. Corsarr handles the infrastructure.",
      body: "A media server usually requires several tools, paths, and credentials. Corsarr turns that setup into one clear journey.",
      items: ["No Docker in the main experience", "No YAML files to edit", "No wiring every app by hand"],
    },
    steps: {
      eyebrow: "HOW IT WORKS",
      title: "From an empty computer to an organized library.",
      body: "You choose what you want to use. Corsarr prepares the rest while keeping every app available through its original interface.",
      items: [
        {
          title: "Choose your storage",
          body: "Point to a folder. Corsarr checks free space, write access, and efficiency before creating anything.",
        },
        {
          title: "Shape your experience",
          body: "Choose movies, shows, subtitles, requests, and streaming — or start with the recommended setup.",
        },
        {
          title: "Install in one flow",
          body: "Corsarr installs the apps, creates consistent paths, and connects supported integrations.",
        },
        {
          title: "Watch and operate",
          body: "Open, update, restart, or remove apps without losing your library or configuration.",
        },
      ],
    },
    apps: {
      eyebrow: "APPS THAT ALREADY WORK TOGETHER",
      title: "Great tools, connected for you.",
      body: "Corsarr does not replace these projects. It provides a simple, careful path to install and connect them.",
      trademark:
        "Names and logos belong to their respective projects. Corsarr is not affiliated with or endorsed by them.",
    },
    safety: {
      eyebrow: "LOCAL BY DEFAULT",
      title: "Your computer. Your files. Your choices.",
      body: "Simplicity should not hide important decisions. Corsarr reduces technical operation while keeping data and consent under your control.",
      cards: [
        {
          title: "Private by default",
          body: "Admin panels are available only on this computer. Jellyfin access on your local network is a separate choice.",
        },
        {
          title: "Data preserved",
          body: "Removing an app does not delete your media. Configuration and library have distinct, explicit actions.",
        },
        {
          title: "Reviewed components",
          body: "The catalog uses identified versions and validated runtime contracts before touching existing services.",
        },
      ],
    },
    responsible: {
      title: "Made for your legitimate library.",
      body: "Corsarr does not include movies, shows, indexers, or third-party accounts. Use it only with content and sources you have the right to access, following applicable laws and terms.",
    },
    cli: {
      eyebrow: "PREFER THE TERMINAL?",
      title: "The same flag, with advanced control.",
      body: "The CLI remains available for servers, home labs, and people who want to generate and review their Docker Compose configuration.",
      command: "corsarr generate",
      action: "CLI documentation",
    },
    downloads: {
      eyebrow: "CHOOSE YOUR PLATFORM",
      title: "Ready to begin?",
      body: "Desktop builds are published alongside Corsarr releases. Choose the platform for this computer.",
      platforms: [
        {
          name: "macOS",
          detail: "Apple Silicon + Intel",
          action: "Download for macOS",
          href: `${releases}/latest/download/corsarr_desktop_darwin_universal.zip`,
        },
        {
          name: "Windows",
          detail: "Windows 10/11 · x64",
          action: "Download for Windows",
          href: `${releases}/latest/download/corsarr_desktop_windows_amd64.zip`,
        },
        {
          name: "Linux",
          detail: "Linux · x64",
          action: "Download for Linux",
          href: `${releases}/latest/download/corsarr_desktop_linux_amd64.tar.gz`,
        },
      ],
      notice: "Early builds may be unsigned. Check the release notes before installing.",
    },
    faq: {
      eyebrow: "FREQUENTLY ASKED QUESTIONS",
      title: "Before setting sail.",
      items: [
        {
          question: "Do I need to know Docker?",
          answer:
            "Not to use Corsarr Desktop. The runtime remains visible in terms, licenses, and technical details, but the main journey is about apps, storage, and media.",
        },
        {
          question: "Does Corsarr provide movies or shows?",
          answer:
            "No. Corsarr installs and connects software. It does not provide media, indexers, accounts, VPNs, or download sources. You are responsible for using only permitted content and services.",
        },
        {
          question: "Can I watch on my TV?",
          answer:
            "Yes. When Jellyfin is selected, Corsarr can expose only streaming to devices on the same local network. Admin panels remain private.",
        },
        {
          question: "What happens when I remove an app?",
          answer:
            "The container is removed, while configuration, downloads, and library are preserved. Deleting configuration is a separate action and never includes shared media.",
        },
        {
          question: "Are Windows and Linux supported?",
          answer:
            "Builds are planned for all three platforms. Full onboarding starts on macOS; check each release note for the current Windows and Linux status.",
        },
        {
          question: "Is there still a CLI?",
          answer:
            "Yes. The CLI remains available for advanced users and can generate the stack Docker Compose configuration.",
        },
      ],
    },
    footer: {
      body: "Navigate the high seas of media automation with simplicity and local control.",
      github: "Source on GitHub",
      licenses: "Licenses and credits",
      copyright: "Independent open-source project.",
    },
  },
};

const applications = [
  { name: "Jellyfin", slug: "jellyfin", role: { "pt-BR": "Assista", en: "Watch" } },
  { name: "Sonarr", slug: "sonarr", role: { "pt-BR": "Séries", en: "Shows" } },
  { name: "Radarr", slug: "radarr", role: { "pt-BR": "Filmes", en: "Movies" } },
  { name: "Prowlarr", slug: "prowlarr", role: { "pt-BR": "Buscas", en: "Search" } },
  { name: "qBittorrent", slug: "qbittorrent", role: { "pt-BR": "Downloads", en: "Downloads" } },
  { name: "Bazarr", slug: "bazarr", role: { "pt-BR": "Legendas", en: "Subtitles" } },
  { name: "Seerr", slug: "seerr", role: { "pt-BR": "Pedidos", en: "Requests" } },
];

const locale = document.documentElement.dataset.locale === "en" ? "en" : "pt-BR";
const t = copy[locale];
const alternatePath = locale === "en" ? "/" : "/en/";

const root = document.querySelector<HTMLDivElement>("#app");
if (!root) throw new Error("Corsarr website root was not found.");

const icon = (name: "apple" | "windows" | "linux" | "terminal" | "arrow") =>
  `<span class="icon icon-${name}" aria-hidden="true"></span>`;

root.innerHTML = `
  <header class="site-header">
    <a class="brand" href="${locale === "en" ? "/en/" : "/"}" aria-label="Corsarr">
      <img src="/corsarr-logo.png" alt="" width="48" height="48">
      <span><strong>Corsarr</strong><small>DESKTOP + CLI</small></span>
    </a>
    <button class="nav-toggle" type="button" aria-expanded="false" aria-controls="site-nav"><span></span><span></span><span></span><span class="sr-only">Menu</span></button>
    <nav id="site-nav" class="site-nav" aria-label="${locale === "en" ? "Main navigation" : "Navegação principal"}">
      <a href="#how">${t.nav.how}</a>
      <a href="#apps">${t.nav.apps}</a>
      <a href="#safety">${t.nav.safety}</a>
      <a href="#faq">${t.nav.faq}</a>
      <a href="${github}" target="_blank" rel="noreferrer">${t.nav.github}</a>
      <a class="language-link" href="${alternatePath}" lang="${locale === "en" ? "pt-BR" : "en"}">${t.nav.language}</a>
    </nav>
  </header>

  <main id="main">
    <section class="hero section-shell">
      <div class="hero-copy">
        <p class="eyebrow">${t.hero.eyebrow}</p>
        <h1>${t.hero.title}<br><em>${t.hero.accent}</em></h1>
        <p class="hero-body">${t.hero.body}</p>
        <div class="hero-actions">
          <a class="primary-button" href="#download">${t.hero.download} ${icon("arrow")}</a>
          <a class="text-button" href="#cli">${icon("terminal")} ${t.hero.cli}</a>
        </div>
        <p class="hero-note">${t.hero.note}</p>
      </div>
      <div class="hero-radar" aria-hidden="true"><span></span><span></span><i></i><b>C</b><small></small></div>
    </section>

    <section class="product-preview section-shell" aria-labelledby="preview-title">
      <div class="window-frame">
        <div class="window-bar"><span></span><span></span><span></span><strong>Corsarr</strong></div>
        <picture>
          <img src="/screenshots/corsarr-dashboard.jpg" alt="${t.preview.caption}" width="1117" height="768">
        </picture>
        <div class="preview-fallback">
          <p class="eyebrow">${t.preview.eyebrow}</p>
          <h2 id="preview-title">${t.preview.title}<br><em>${t.preview.status}</em></h2>
          <p>${t.preview.services}</p>
        </div>
      </div>
      <p class="image-caption">${t.preview.caption}</p>
      <figure class="onboarding-shot">
        <img src="/screenshots/corsarr-onboarding.jpg" alt="${t.preview.onboardingCaption}" width="1117" height="768" loading="lazy">
        <figcaption>${t.preview.onboardingCaption}</figcaption>
      </figure>
    </section>

    <section class="problem section-shell">
      <div>
        <p class="eyebrow">${t.problem.eyebrow}</p>
        <h2>${t.problem.title}</h2>
      </div>
      <div>
        <p>${t.problem.body}</p>
        <ul>${t.problem.items.map((item) => `<li><span>✓</span>${item}</li>`).join("")}</ul>
      </div>
    </section>

    <section id="how" class="steps section-shell">
      <div class="section-heading">
        <p class="eyebrow">${t.steps.eyebrow}</p>
        <h2>${t.steps.title}</h2>
        <p>${t.steps.body}</p>
      </div>
      <ol class="step-grid">
        ${t.steps.items.map((item, index) => `<li><span>0${index + 1}</span><h3>${item.title}</h3><p>${item.body}</p></li>`).join("")}
      </ol>
    </section>

    <section id="apps" class="apps section-shell">
      <div class="section-heading centered">
        <p class="eyebrow">${t.apps.eyebrow}</p>
        <h2>${t.apps.title}</h2>
        <p>${t.apps.body}</p>
      </div>
      <div class="app-orbit">
        <div class="app-grid">
          ${applications.map((app) => `<article class="app-card"><img src="/apps/${app.slug}.png" alt="Logo ${app.name}" width="72" height="72"><div><h3>${app.name}</h3><p>${app.role[locale]}</p></div></article>`).join("")}
        </div>
      </div>
      <p class="trademark-note">${t.apps.trademark}</p>
    </section>

    <section id="safety" class="safety">
      <div class="section-shell">
        <div class="section-heading safety-heading">
          <p class="eyebrow">${t.safety.eyebrow}</p>
          <h2>${t.safety.title}</h2>
          <p>${t.safety.body}</p>
        </div>
        <div class="safety-grid">
          ${t.safety.cards.map((card, index) => `<article><span class="safety-symbol">${["◎", "▱", "◇"][index]}</span><h3>${card.title}</h3><p>${card.body}</p></article>`).join("")}
        </div>
        <aside class="responsible-note"><span>i</span><div><h3>${t.responsible.title}</h3><p>${t.responsible.body}</p></div></aside>
      </div>
    </section>

    <section id="cli" class="cli section-shell">
      <div>
        <p class="eyebrow">${t.cli.eyebrow}</p>
        <h2>${t.cli.title}</h2>
        <p>${t.cli.body}</p>
        <a class="text-button" href="${github}#quick-start" target="_blank" rel="noreferrer">${t.cli.action} ${icon("arrow")}</a>
      </div>
      <div class="terminal-card" aria-label="Terminal"><div><span></span><span></span><span></span></div><code><b>$</b> ${t.cli.command}<i></i></code></div>
    </section>

    <section id="download" class="downloads section-shell">
      <div class="section-heading centered">
        <p class="eyebrow">${t.downloads.eyebrow}</p>
        <h2>${t.downloads.title}</h2>
        <p>${t.downloads.body}</p>
      </div>
      <div class="download-grid">
        ${t.downloads.platforms.map((platform, index) => `<article><span class="platform-icon">${icon((["apple", "windows", "linux"] as const)[index] ?? "linux")}</span><h3>${platform.name}</h3><p>${platform.detail}</p><a href="${platform.href}" rel="noreferrer">${platform.action} ${icon("arrow")}</a></article>`).join("")}
      </div>
      <p class="download-notice">${t.downloads.notice} <a href="${releases}">${locale === "en" ? "View all releases." : "Ver todas as releases."}</a></p>
    </section>

    <section id="faq" class="faq section-shell">
      <div class="section-heading">
        <p class="eyebrow">${t.faq.eyebrow}</p>
        <h2>${t.faq.title}</h2>
      </div>
      <div class="faq-list">
        ${t.faq.items.map((item) => `<details><summary>${item.question}<span></span></summary><p>${item.answer}</p></details>`).join("")}
      </div>
    </section>
  </main>

  <footer class="site-footer section-shell">
    <div class="footer-brand"><img src="/corsarr-logo.png" alt="" width="58" height="58"><div><strong>Corsarr</strong><p>${t.footer.body}</p></div></div>
    <div class="footer-links"><a href="${github}">${t.footer.github}</a><a href="${github}/blob/main/website/THIRD_PARTY_ASSETS.md">${t.footer.licenses}</a></div>
    <div class="footer-bottom"><span>© ${new Date().getFullYear()} Corsarr</span><span>${t.footer.copyright}</span></div>
  </footer>
`;

const navigationToggle = document.querySelector<HTMLButtonElement>(".nav-toggle");
const navigation = document.querySelector<HTMLElement>(".site-nav");
navigationToggle?.addEventListener("click", () => {
  const expanded = navigationToggle.getAttribute("aria-expanded") === "true";
  navigationToggle.setAttribute("aria-expanded", String(!expanded));
  navigation?.classList.toggle("open", !expanded);
});

navigation?.addEventListener("click", (event) => {
  if ((event.target as HTMLElement).closest("a")) {
    navigationToggle?.setAttribute("aria-expanded", "false");
    navigation.classList.remove("open");
  }
});

const screenshot = document.querySelector<HTMLImageElement>(".window-frame img");
const revealScreenshot = () => screenshot?.closest(".window-frame")?.classList.add("has-image");
if (screenshot?.complete) revealScreenshot();
else screenshot?.addEventListener("load", revealScreenshot, { once: true });
