export const gridLayout = [
  ["BindMove", "BindStop", "BindHold", "BindAttack"],
  ["BindPatrol", "SkillSlot5", "SkillSlot6", "BindOpenHeroSkills"],
  ["SkillSlot1", "SkillSlot2", "SkillSlot3", "SkillSlot4"],
];

export const allButtons = {
  BindMove: { id: "BindMove", icon: "BTNMove.png" },
  BindStop: { id: "BindStop", icon: "BTNStop.png" },
  BindHold: { id: "BindHold", icon: "BTNHold.png" },
  BindPatrol: { id: "BindPatrol", icon: "BTNPatrol.png" },
  BindAttack: { id: "BindAttack", icon: "BTNAttack.png" },
  BindOpenHeroSkills: { id: "BindOpenHeroSkills", icon: "BTNSkillz.png" },

  SkillSlot1: { id: "SkillSlot1", icon: "BTNSkillSlot.png" },
  SkillSlot2: { id: "SkillSlot2", icon: "BTNSkillSlot.png" },
  SkillSlot3: { id: "SkillSlot3", icon: "BTNSkillSlot.png" },
  SkillSlot4: { id: "SkillSlot4", icon: "BTNSkillSlot.png" },
  SkillSlot5: { id: "SkillSlot5", icon: "BTNSkillSlot.png" },
  SkillSlot6: { id: "SkillSlot6", icon: "BTNSkillSlot.png" },

  QuickCastSlot1: { id: "QuickCastSlot1", icon: "BTNSkillSlot.png" },
  QuickCastSlot2: { id: "QuickCastSlot2", icon: "BTNSkillSlot.png" },
  QuickCastSlot3: { id: "QuickCastSlot3", icon: "BTNSkillSlot.png" },
  QuickCastSlot4: { id: "QuickCastSlot4", icon: "BTNSkillSlot.png" },
  QuickCastSlot5: { id: "QuickCastSlot5", icon: "BTNSkillSlot.png" },
  QuickCastSlot6: { id: "QuickCastSlot6", icon: "BTNSkillSlot.png" },

  ASkillSlot1: { id: "ASkillSlot1", icon: "BTNSkillSlot.png" },
  ASkillSlot2: { id: "ASkillSlot2", icon: "BTNSkillSlot.png" },
  ASkillSlot3: { id: "ASkillSlot3", icon: "BTNSkillSlot.png" },
  ASkillSlot4: { id: "ASkillSlot4", icon: "BTNSkillSlot.png" },
  ASkillSlot5: { id: "ASkillSlot5", icon: "BTNSkillSlot.png" },
  ASkillSlot6: { id: "ASkillSlot6", icon: "BTNSkillSlot.png" },

  ExtraSlot1: {
    id: "ExtraSlot1",
    icon: "BTNAttribute.png",
    disabled: true,
  },
  ExtraSlot2: { id: "ExtraSlot2", icon: "BTNSkillSlot.png" },
  ExtraSlot3: { id: "ExtraSlot3", icon: "BTNSkillSlot.png" },
  ExtraSlot5: { id: "ExtraSlot5", icon: "BTNSkillSlot.png" },

  QuickcastExtraSlot2: {
    id: "QuickCastExtraSlot2",
    icon: "BTNSkillSlot.png",
  },
  QuickcastExtraSlot3: {
    id: "QuickCastExtraSlot3",
    icon: "BTNSkillSlot.png",
  },
  QuickcastExtraSlot5: {
    id: "QuickCastExtraSlot5",
    icon: "BTNSkillSlot.png",
  },
  QuickcastExtraSlot6: {
    id: "QuickCastExtraSlot6",
    icon: "BTNSkillSlot.png",
  },

  AutocastExtraSlot2: {
    id: "AutoCastExtraSlot2",
    icon: "BTNSkillSlot.png",
  },
  AutocastExtraSlot3: {
    id: "AutoCastExtraSlot3",
    icon: "BTNSkillSlot.png",
  },
  AutocastExtraSlot5: {
    id: "AutoCastExtraSlot5",
    icon: "BTNSkillSlot.png",
  },
  AutocastExtraSlot6: {
    id: "AutocastExtraSlot6",
    icon: "BTNSkillSlot.png",
  },
};

export const extraRules = {
  BindMove: { any: "ExtraSlot1" },
  BindStop: {
    cast: "ExtraSlot2",
    quickcast: "QuickcastExtraSlot2",
    autocast: "AutocastExtraSlot2",
  },
  BindHold: {
    cast: "ExtraSlot3",
    quickcast: "QuickcastExtraSlot3",
    autocast: "AutocastExtraSlot3",
  },
  BindPatrol: {
    cast: "ExtraSlot5",
    quickcast: "QuickcastExtraSlot5",
    autocast: "AutocastExtraSlot5",
  },
};

export const skillRules = {
  cast: (id) => allButtons[id],
  quickcast: (id) => allButtons[`QuickCast${id.substring(5)}`],
  autocast: (id) => allButtons[`A${id}`],
};
