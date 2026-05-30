// Domain constants for Super Mega Baseball player attributes.
// All trait, position, chemistry, and hand constants reflect SMB4 only.

export const BATTING_POSITIONS = ['C', '1B', '2B', '3B', 'SS', 'LF', 'CF', 'RF', 'DH'] as const
export const PITCHING_ROLES = ['SP', 'RP', 'SP/RP', 'CL'] as const
export const BAT_HANDS = ['L', 'R', 'S'] as const
export const THROW_HANDS = ['L', 'R'] as const
export const CHEMISTRY_TYPES = ['Competitive', 'Spirited', 'Disciplined', 'Scholarly', 'Crafty'] as const

// All 80 SMB4 traits, sorted alphabetically.
export const SMB4_TRAITS = [
  'Ace Exterminator',
  'Bad Ball Hitter',
  'Bad Jumps',
  'Base Jogger',
  'Base Rounder',
  'BB Prone',
  'Big Hack',
  'Bunter',
  'Butter Fingers',
  'Cannon Arm',
  'Choker',
  'Clutch',
  'CON vs LHP',
  'CON vs RHP',
  'Composed',
  'Consistent',
  'Crossed Up',
  'Distractor',
  'Dive Wizard',
  'Durable',
  'Easy Jumps',
  'Easy Target',
  'Elite 2F',
  'Elite 4F',
  'Elite CB',
  'Elite CF',
  'Elite CH',
  'Elite FK',
  'Elite SB',
  'Elite SL',
  'Falls Behind',
  'Fastball Hitter',
  'First Pitch Prayer',
  'First Pitch Slayer',
  'Gets Ahead',
  'High Pitch',
  'Injury Prone',
  'Inside Pitch',
  'K Collector',
  'K Neglector',
  'Little Hack',
  'Low Pitch',
  'Magic Hands',
  'Meltdown',
  'Metal Head',
  'Mind Gamer',
  'Noodle Arm',
  'Off-speed Hitter',
  'Outside Pitch',
  'Pick Officer',
  'Pinch Perfect',
  'POW vs LHP',
  'POW vs RHP',
  'Rally Starter',
  'Rally Stopper',
  'RBI Hero',
  'RBI Zero',
  'Reverse Splits',
  'Sign Stealer',
  'Slow Poke',
  'Specialist',
  'Sprinter',
  'Stealer',
  'Stimulated',
  'Surrounded',
  'Tough Out',
  'Two Way (C)',
  'Two Way (IF)',
  'Two Way (OF)',
  'Utility',
  'Volatile',
  'Whiffer',
  'Wild Thing',
  'Wild Thrower',
  'Workhorse',
] as const

export type BattingPosition = (typeof BATTING_POSITIONS)[number]
export type PitchingRole = (typeof PITCHING_ROLES)[number]
export type BatHand = (typeof BAT_HANDS)[number]
export type ThrowHand = (typeof THROW_HANDS)[number]
export type ChemistryType = (typeof CHEMISTRY_TYPES)[number]
export type SMB4Trait = (typeof SMB4_TRAITS)[number]
