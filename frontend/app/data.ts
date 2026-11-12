// Shared Priyosha content — used by every page.
// The Go api-gateway. Set NEXT_PUBLIC_API_URL in production, e.g.
// https://api.priyoshacharity.com
export const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export type Stat = {
  id: string;
  title: string;
  figure: string;
  detail: string;
  source: string;
};

export type Program = {
  id: string;
  title: string;
  description: string;
  focus: string;
};

// Fallback copy of the content-service data, used only if the API cannot be
// reached (so pages still show the correct figures). The API is the source
// of truth when it is up — edit figures in backend/content-service.
export const FALLBACK_WARNING =
  "These figures cannot be added together: hunger and homelessness can affect the same people, and animal population estimates do not measure suffering.";

export const FALLBACK_STATS: Stat[] = [
  {
    id: "food-insecurity",
    title: "People lacking enough food",
    figure: "18.1 million — 1 crore 81 lakh",
    detail:
      "Projected to face crisis-level food insecurity or worse during September–December 2026, within the areas assessed.",
    source: "Integrated Food Security Phase Classification (IPC)",
  },
  {
    id: "sleeping-in-public",
    title: "People sleeping in public places",
    figure: "22,185",
    detail:
      "Counted in the 2022 census. This narrow category covers people sleeping on streets, at stations and similar places. It does not capture everyone without secure housing, and it is an older count.",
    source: "Bangladesh Bureau of Statistics, 2022 census (nsds.bbs.gov.bd)",
  },
  {
    id: "dogs",
    title: "Dogs",
    figure: "About 16.7 lakh dogs",
    detail:
      "Estimated in a 2024 study. This estimates the dog population; it does not establish how many are homeless or hungry.",
    source: "2024 study (PMC)",
  },
  {
    id: "animals-overall",
    title: "Hungry or homeless animals overall, including cats",
    figure: "No reliable nationwide total found",
    detail:
      "Research identifies gaps in Bangladesh's stray-animal population data. Giving an exact suffering count would be guesswork.",
    source: "pmc.ncbi.nlm.nih.gov",
  },
];

export const FALLBACK_PROGRAMS: Program[] = [
  {
    id: "food-for-people",
    title: "Food for people",
    description:
      "Food support for people facing crisis-level food insecurity in Bangladesh — meals and food packs for families who do not have enough to eat.",
    focus: "People",
  },
  {
    id: "support-sleeping-rough",
    title: "Support for people sleeping rough",
    description:
      "Help for people sleeping in public places — on streets, at stations and similar places — including shelter support, essentials and a path to more secure housing.",
    focus: "People",
  },
  {
    id: "care-for-animals",
    title: "Food and care for cats and dogs",
    description:
      "Food, shelter and care for stray cats and dogs who suffer the most — feeding, basic veterinary care and rescue where we can reach them.",
    focus: "Animals — cats and dogs",
  },
];
