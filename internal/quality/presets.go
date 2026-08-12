package quality

type PresetID string

// PresetCatalogVersion identifies the exact Corsarr mapping from the
// user-facing presets to pinned TRaSH Guides profile identifiers.
const PresetCatalogVersion = "2026-08-12.1"

const (
	PresetEconomy       PresetID = "economy"
	PresetBalanced1080p PresetID = "balanced-1080p"
	PresetHigh1080p     PresetID = "high-1080p"
	Preset4KHDR         PresetID = "4k-hdr"
	PresetUnmanaged     PresetID = "unmanaged"
)

type Preset struct {
	ID            PresetID `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Summary       string   `json:"summary"`
	RadarrTrashID string   `json:"-"`
	SonarrTrashID string   `json:"-"`
}

var presets = []Preset{
	{
		ID: PresetEconomy, Name: "Econômico em armazenamento",
		Description:   "Prioriza versões WEB em 1080p e evita arquivos Blu-ray ou Remux maiores.",
		Summary:       "Cria o perfil Corsarr econômico, aplica os formatos recomendados e limites WEB 1080p do guia e padroniza nomes de pastas e arquivos.",
		RadarrTrashID: "e8c5acb741363a0dbda67d3978f4912f",
		SonarrTrashID: "72dae194fc92bf828f32cde7744e51a1",
	},
	{
		ID: PresetBalanced1080p, Name: "Equilibrado 1080p",
		Description:   "Equilibra compatibilidade, disponibilidade e qualidade até 1080p.",
		Summary:       "Cria o perfil Corsarr equilibrado, aplica formatos recomendados, limites para WEB e Blu-ray até 1080p e nomes de mídia padronizados.",
		RadarrTrashID: "d1d67249d3890e49bc12e275d989a7e9",
		SonarrTrashID: "9d142234e45d6143785ac55f5a9e8dc9",
	},
	{
		ID: PresetHigh1080p, Name: "Alta qualidade 1080p",
		Description:   "Aceita arquivos Remux 1080p maiores para preservar mais qualidade.",
		Summary:       "Cria o perfil Corsarr de alta qualidade, aplica formatos recomendados, limites para Remux e WEB 1080p e nomes de mídia padronizados.",
		RadarrTrashID: "9ca12ea80aa55ef916e3751f4b874151",
		SonarrTrashID: "fe9470e577c300a5ad9a3274f6d1cdf2",
	},
	{
		ID: Preset4KHDR, Name: "4K e HDR",
		Description:   "Prioriza conteúdo 2160p e formatos HDR para dispositivos compatíveis.",
		Summary:       "Cria o perfil Corsarr 4K, aplica formatos HDR e demais recomendações compatíveis, limites 2160p e nomes de mídia padronizados.",
		RadarrTrashID: "64fb5f9858489bdac2af690e27c8f42f",
		SonarrTrashID: "d1498e7d189fbe6c7110ceaabb7473e6",
	},
	{
		ID: PresetUnmanaged, Name: "Configurar por conta própria",
		Description: "O Corsarr não altera perfis, formatos personalizados ou limites de qualidade.",
		Summary:     "O Corsarr também não altera nomes de mídia; Radarr, Sonarr ou outra ferramenta permanecem sob sua administração.",
	},
}

func Presets() []Preset {
	return append([]Preset(nil), presets...)
}

func FindPreset(id PresetID) (Preset, bool) {
	for _, preset := range presets {
		if preset.ID == id {
			return preset, true
		}
	}
	return Preset{}, false
}

func ValidPreset(id string) bool {
	_, found := FindPreset(PresetID(id))
	return found
}

func ProfileName(id PresetID) string {
	preset, found := FindPreset(id)
	if !found || id == PresetUnmanaged {
		return ""
	}
	return "Corsarr - " + preset.Name
}
