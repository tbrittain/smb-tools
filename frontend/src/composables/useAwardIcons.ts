const AWARD_ICONS: Record<string, string> = {
  'League Champion': '🏆',
  'Conference Champion': '🥈',
}

// Maps award originalName → importance tier (matches the awards table seed data).
const AWARD_IMPORTANCE: Record<string, number> = {
  MVP: 0,
  'Triple Crown (Batting)': 0,
  'Triple Crown (Pitching)': 0,
  'Cy Young': 1,
  'Silver Slugger': 1,
  ROY: 1,
  'Gold Glove': 2,
  'Playoff MVP': 2,
  'Championship MVP': 2,
  'Batting Title': 3,
  'Home Run Title': 3,
  'RBI Title': 3,
  'ERA Title': 3,
  'Wins Title': 3,
  'Strikeouts Title': 3,
  'All-Star': 4,
  'MVP-2': 5,
  'MVP-3': 5,
  'MVP-4': 5,
  'MVP-5': 5,
  'Cy Young-2': 5,
  'Cy Young-3': 5,
  'Cy Young-4': 5,
  'Cy Young-5': 5,
  'ROY-2': 5,
  'ROY-3': 5,
  'ROY-4': 5,
  'ROY-5': 5,
  'League Champion': 0,
  'Conference Champion': 1,
}

export function getAwardIcon(originalName: string): string {
  return AWARD_ICONS[originalName] ?? ''
}

export function getAwardImportance(originalName: string): number {
  return AWARD_IMPORTANCE[originalName] ?? 4
}
