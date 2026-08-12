const en = {
  'quality.economy.name': 'Storage efficient',
  'quality.economy.description':
    'Prioritizes 1080p WEB releases and avoids larger Blu-ray or Remux files.',
  'quality.economy.summary':
    'Creates the efficient Corsarr profile, applies recommended formats and 1080p WEB limits, and standardizes folder and file names.',
  'quality.balanced-1080p.name': 'Balanced 1080p',
  'quality.balanced-1080p.description':
    'Balances compatibility, availability, and quality up to 1080p.',
  'quality.balanced-1080p.summary':
    'Creates the balanced Corsarr profile, applies recommended formats, WEB and Blu-ray limits up to 1080p, and standardized media names.',
  'quality.high-1080p.name': 'High quality 1080p',
  'quality.high-1080p.description': 'Accepts larger 1080p Remux files to preserve more quality.',
  'quality.high-1080p.summary':
    'Creates the high-quality Corsarr profile, applies recommended formats, Remux and 1080p WEB limits, and standardized media names.',
  'quality.4k-hdr.name': '4K and HDR',
  'quality.4k-hdr.description': 'Prioritizes 2160p content and HDR formats for compatible devices.',
  'quality.4k-hdr.summary':
    'Creates the Corsarr 4K profile, applies HDR and compatible recommendations, 2160p limits, and standardized media names.',
  'quality.unmanaged.name': 'Configure it myself',
  'quality.unmanaged.description':
    'Corsarr does not change profiles, custom formats, or quality limits.',
  'quality.unmanaged.summary':
    'Corsarr also does not change media names; Radarr, Sonarr, or another tool remains under your administration.',
} as const;
type QualityCatalog = Record<keyof typeof en, string>;
const es: QualityCatalog = {
  'quality.economy.name': 'Ahorro de almacenamiento',
  'quality.economy.description':
    'Prioriza versiones WEB en 1080p y evita archivos Blu-ray o Remux más grandes.',
  'quality.economy.summary':
    'Crea el perfil Corsarr eficiente, aplica formatos recomendados y límites WEB 1080p y estandariza nombres de carpetas y archivos.',
  'quality.balanced-1080p.name': 'Equilibrado 1080p',
  'quality.balanced-1080p.description':
    'Equilibra compatibilidad, disponibilidad y calidad hasta 1080p.',
  'quality.balanced-1080p.summary':
    'Crea el perfil Corsarr equilibrado, aplica formatos recomendados, límites WEB y Blu-ray hasta 1080p y nombres multimedia estandarizados.',
  'quality.high-1080p.name': 'Alta calidad 1080p',
  'quality.high-1080p.description':
    'Acepta archivos Remux 1080p más grandes para conservar más calidad.',
  'quality.high-1080p.summary':
    'Crea el perfil Corsarr de alta calidad, aplica formatos recomendados, límites Remux y WEB 1080p y nombres multimedia estandarizados.',
  'quality.4k-hdr.name': '4K y HDR',
  'quality.4k-hdr.description':
    'Prioriza contenido 2160p y formatos HDR para dispositivos compatibles.',
  'quality.4k-hdr.summary':
    'Crea el perfil Corsarr 4K, aplica HDR y recomendaciones compatibles, límites 2160p y nombres multimedia estandarizados.',
  'quality.unmanaged.name': 'Configurar por mi cuenta',
  'quality.unmanaged.description':
    'Corsarr no modifica perfiles, formatos personalizados ni límites de calidad.',
  'quality.unmanaged.summary':
    'Corsarr tampoco modifica los nombres multimedia; Radarr, Sonarr u otra herramienta quedan bajo tu administración.',
};
const ptBR: QualityCatalog = {
  'quality.economy.name': 'Econômico em armazenamento',
  'quality.economy.description':
    'Prioriza versões WEB em 1080p e evita arquivos Blu-ray ou Remux maiores.',
  'quality.economy.summary':
    'Cria o perfil Corsarr econômico, aplica os formatos recomendados e limites WEB 1080p do guia e padroniza nomes de pastas e arquivos.',
  'quality.balanced-1080p.name': 'Equilibrado 1080p',
  'quality.balanced-1080p.description':
    'Equilibra compatibilidade, disponibilidade e qualidade até 1080p.',
  'quality.balanced-1080p.summary':
    'Cria o perfil Corsarr equilibrado, aplica formatos recomendados, limites para WEB e Blu-ray até 1080p e nomes de mídia padronizados.',
  'quality.high-1080p.name': 'Alta qualidade 1080p',
  'quality.high-1080p.description':
    'Aceita arquivos Remux 1080p maiores para preservar mais qualidade.',
  'quality.high-1080p.summary':
    'Cria o perfil Corsarr de alta qualidade, aplica formatos recomendados, limites para Remux e WEB 1080p e nomes de mídia padronizados.',
  'quality.4k-hdr.name': '4K e HDR',
  'quality.4k-hdr.description':
    'Prioriza conteúdo 2160p e formatos HDR para dispositivos compatíveis.',
  'quality.4k-hdr.summary':
    'Cria o perfil Corsarr 4K, aplica formatos HDR e demais recomendações compatíveis, limites 2160p e nomes de mídia padronizados.',
  'quality.unmanaged.name': 'Configurar por conta própria',
  'quality.unmanaged.description':
    'O Corsarr não altera perfis, formatos personalizados ou limites de qualidade.',
  'quality.unmanaged.summary':
    'O Corsarr também não altera nomes de mídia; Radarr, Sonarr ou outra ferramenta permanecem sob sua administração.',
};
const it: QualityCatalog = {
  'quality.economy.name': 'Risparmio di spazio',
  'quality.economy.description':
    'Privilegia le versioni WEB 1080p ed evita file Blu-ray o Remux più grandi.',
  'quality.economy.summary':
    'Crea il profilo Corsarr efficiente, applica i formati consigliati e i limiti WEB 1080p e standardizza i nomi di cartelle e file.',
  'quality.balanced-1080p.name': 'Bilanciato 1080p',
  'quality.balanced-1080p.description':
    'Bilancia compatibilità, disponibilità e qualità fino a 1080p.',
  'quality.balanced-1080p.summary':
    'Crea il profilo Corsarr bilanciato, applica formati consigliati, limiti WEB e Blu-ray fino a 1080p e nomi multimediali standardizzati.',
  'quality.high-1080p.name': 'Alta qualità 1080p',
  'quality.high-1080p.description':
    'Accetta file Remux 1080p più grandi per preservare maggiore qualità.',
  'quality.high-1080p.summary':
    'Crea il profilo Corsarr di alta qualità, applica formati consigliati, limiti Remux e WEB 1080p e nomi multimediali standardizzati.',
  'quality.4k-hdr.name': '4K e HDR',
  'quality.4k-hdr.description':
    'Privilegia contenuti 2160p e formati HDR per i dispositivi compatibili.',
  'quality.4k-hdr.summary':
    'Crea il profilo Corsarr 4K, applica HDR e raccomandazioni compatibili, limiti 2160p e nomi multimediali standardizzati.',
  'quality.unmanaged.name': 'Configura autonomamente',
  'quality.unmanaged.description':
    'Corsarr non modifica profili, formati personalizzati o limiti di qualità.',
  'quality.unmanaged.summary':
    'Corsarr non modifica neppure i nomi multimediali; Radarr, Sonarr o un altro strumento restano sotto la tua amministrazione.',
};
export const qualityMessages = { en, es, 'pt-BR': ptBR, it } as const;
