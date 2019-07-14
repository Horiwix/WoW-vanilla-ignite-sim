// for i in {1..5}; do go run main.go > output/out$i.txt & done

package main

import (
  "time"
  "math"
  "math/rand"
  "sync"
  "log"
  "fmt"
)

var TimeScale float32 = 20 // x times to speed up simulation; more than 50x might ne inaccurate
var SimulationDuration = 91 * time.Second // will be scaled; anything below 20 might be inaccurate
var SimulationRunCount = 5

const (
  LOG_CAST_START bool = false
  LOG_INITIAL_DMG bool = false
  LOG_DOT_DMG bool = false
  LOG_ADD_DEBUFF_STACK bool = false
  LOG_OVERRIDE_DEBUFF_STACK bool = false
  LOG_ADD_DEBUFF_INSTANCE bool = false
  LOG_DEBUFF_DROPPED bool = false
)

var Horiwix = Mage{
  Name: "Horiwix",
  SP: 524+150+36+35+10,
  Crit: 8+5.38+6+1,
  Hit: 8+6,
  DmgScale: 1.1,
  StartDelay: 100 * time.Millisecond,
  Strategy: StrategyCombustOnlyFireball,
}

var Magenificent = Mage{
  Name: "Magenificent",
  SP: 520+150+36+35+10,
  Crit: 7+5.66+6+1,
  Hit: 8+6,
  DmgScale: 1.1,
  StartDelay: 100 * time.Millisecond,
  Strategy: StrategyCombustOnlyFireball,
}

var Jasminbolt = Mage{
  Name: "Jasminbolt",
  SP: 508+150+36+35+10,
  Crit: 5+5.31+6+1,
  Hit: 8+6,
  DmgScale: 1.1,
  StartDelay: 100 * time.Millisecond,
  Strategy: StrategyCombustOnlyFireball,
}

var Worvan = Mage{
  Name: "Worvan",
  SP: 485+150+36+35+10,
  Crit: 5+5.36+6+1,
  Hit: 9+6,
  DmgScale: 1.1,
  Strategy: StrategyScorchCombustOnlyFireball,
}

var Komposter = Mage{
  Name: "Komposter",
  SP: 421+150+36+35+10,
  Crit: 1+6.03+6+1,
  Hit: 4+6,
  DmgScale: 1.1,
  Strategy: StrategyScorchCombustOnlyFireball,
}

var Huell = Mage{
  Name: "Ranajin",
  SP: 334+150+36+35+10,
  Crit: 4+5.63+6+1,
  Hit: 0+6,
  DmgScale: 1.1,
  Strategy: Strategy6ScorchCombustOnlyScorch,
}

var Deluxe = Mage{
  Name: "Deluxe",
  SP: 461+150+35+10,
  Crit: 9.46+6,
  Hit: 0+6,
  DmgScale: 1.1,
  Strategy: StrategyScorchCombustOnlyFireball,
}

var Ranajin = Mage{
  Name: "Ranajin",
  SP: 367+150+35+10,
  Crit: 10+6,
  Hit: 7+6,
  DmgScale: 1.1,
  Strategy: Strategy6ScorchCombustOnlyScorch,
}

var Mages = []Mage{
  Horiwix,
  Magenificent,
  Jasminbolt,
  Worvan,
  Komposter,
  Huell,
}

var TARGET = &Monster{
  Name: "Target Dummy",
  BaseMiss: 17,
  Debuffs: []Debuff{},
}

func StrategyOnlyFireball(mage Mage, start time.Time) {
  StrategyOnlySpell(mage, start, Fireball)
}

func StrategyOnlyScorch(mage Mage, start time.Time) {
  StrategyOnlySpell(mage, start, Scorch)
}

func StrategyCombustOnlyFireball(mage Mage, start time.Time) {
  StrategyCombust3Crits(&mage, start, Fireball)
  StrategyOnlyFireball(mage, start)
}

func StrategyScorchCombustOnlyFireball(mage Mage, start time.Time) {
  StrategySpellUntilHit(mage, start, Scorch)
  StrategyCombustOnlyFireball(mage, start)
}

func StrategyCombustFireballOnlyScorch(mage Mage, start time.Time) {
  StrategyCombust3Crits(&mage, start, Fireball)
  StrategyOnlyScorch(mage, start)
}

func StrategyCombustFireballKeepBigIgnite(mage Mage, start time.Time) {
  StrategyCombust3Crits(&mage, start, Fireball)
  StrategyScorchKeepBigIgnite(mage, start)
}

func Strategy3ScorchCombustOnlyScorch(mage Mage, start time.Time) {
  mage.Cast(Scorch)
  mage.Cast(Scorch)
  mage.Cast(Scorch)
  StrategyCombust3Crits(&mage, start, Scorch)
  StrategyOnlyScorch(mage, start)
}

func Strategy6ScorchCombustOnlyScorch(mage Mage, start time.Time) {
  mage.Cast(Scorch)
  mage.Cast(Scorch)
  mage.Cast(Scorch)
  Strategy3ScorchCombustOnlyScorch(mage, start)
}

func StrategyFireballThenScorch(mage Mage, start time.Time) {
  mage.Cast(Fireball)
  StrategyOnlyScorch(mage, start)
}

func StrategyScorchThenFireball(mage Mage, start time.Time) {
  StrategySpellUntilHit(mage, start, Scorch)
  StrategyOnlyFireball(mage, start)
}

func StrategyScorchCombustFireballKeepScorchUp(mage Mage, start time.Time) {
  StrategySpellUntilHit(mage, start, Scorch)
  StrategyCombust3Crits(&mage, start, Fireball)
  StrategyFireballKeepScorchUp(mage, start)
}

func StrategyFireballKeepScorchUp(mage Mage, start time.Time) {
  lastScorch := time.Now()
  for {
    if time.Now().After(start.Add(ScaleDuration(SimulationDuration))) {
      break
    }

    if time.Now().After(lastScorch.Add(25 * time.Second)) {
      StrategySpellUntilHit(mage, start, Scorch)
      lastScorch = time.Now()
    } else {
      mage.Cast(Fireball)
    }
  }
}

func StrategyScorchFireballKeepScorchUp(mage Mage, start time.Time) {
  StrategySpellUntilHit(mage, start, Scorch)
  StrategyFireballKeepScorchUp(mage, start)
}

func StrategyFireballKeepSoloScorchUp(mage Mage, start time.Time) {
  for i := 0; i < 5; i++ {
    StrategySpellUntilHit(mage, start, Scorch)
  }
  
  StrategyFireballKeepScorchUp(mage, start)
}

func StrategyCombust3Crits(mage *Mage, start time.Time, spell Spell) {
  baseCrit := mage.Crit
  
  for i := 0; i < 3; i++ {
    mage.Crit += 10
    result := mage.Cast(spell)
    if result != CRIT {
      i--
    }
  }
  
  mage.Crit = baseCrit
}

func StrategySpellUntilHit(mage Mage, start time.Time, spell Spell) {
  for {
    result := mage.Cast(spell)
    if result != MISS {
      break
    }
  }
}

func StrategyOnlySpell(mage Mage, start time.Time, spell Spell) {
  for {
    if time.Now().After(start.Add(ScaleDuration(SimulationDuration))) {
      break
    }

    mage.Cast(spell)
  }
}

func StrategyScorchKeepBigIgnite(mage Mage, start time.Time) {
  for {
    if time.Now().After(start.Add(ScaleDuration(SimulationDuration))) {
      break
    }

    if CheckTargetIgniteDamage() <= 1500 {
      mage.Cast(Fireball)
    } else {
      mage.Cast(Scorch)
    }
  }
}

func CheckTargetIgniteDamage() float32 {
  var totalIgniteDmg float32
  for _, debuff := range TARGET.Debuffs {
    if debuff.Name == Ignite.Name {
      for _, stack := range debuff.Stacks {
        totalIgniteDmg += stack.TotalDmg
      }
    }
  }
  
  return totalIgniteDmg
}

// -------------------------------------------------------------------------------

func (mage Mage) Cast(spell Spell) AttackResult {
  if LOG_CAST_START {
    log.Printf("[%s] casting %s...\n", mage.Name, spell.Name)
  }
  time.Sleep(ScaleDuration(spell.CastTime))
  return TARGET.Hit(mage, spell)
}

func (mage Mage) Play() {
  defer wg.Done()

  mage.Strategy(mage, time.Now())
}

func (mob Monster) GetPlusDmg() float32 {
  var plusDmg float32
  for _, debuff := range mob.Debuffs {
    for _, stack := range debuff.Stacks {
      plusDmg += stack.PlusDmg
    }
  }

  return plusDmg
}

func (mob Monster) GetPlusCrit() float32 {
  var plusCrit float32
  for _, debuff := range mob.Debuffs {
    for _, stack := range debuff.Stacks {
      plusCrit += stack.PlusCrit
    }
  }

  return plusCrit
}

func (mob *Monster) ApplyDebuff(mage Mage, effect Effect, attackResult AttackResult) {
  if attackResult < effect.Condition {
    return
  }
  
  lock.Lock()
  defer lock.Unlock()

  if effect.ApplyMask&(DEBUFF_MY_NEW_INSTANCE|DEBUFF_OTHER_NEW_INSTANCE) == 0 {
    // if should NOT create new instance if me and/or other put the same debuff
    var foundDebuff *Debuff

    for i, debuff := range mob.Debuffs {
      if debuff.Name == effect.Name {
        if debuff.ApplierName == mage.Name && effect.ApplyMask&DEBUFF_MY_NEW_STACK != 0 {
          // if debuff is mine and should reapply
          foundDebuff = &mob.Debuffs[i]
          break
        } else if effect.ApplyMask&DEBUFF_OTHER_NEW_STACK != 0 {
          // if debuff is other mage's and should reapply
          foundDebuff = &mob.Debuffs[i]
          break
        }
      }
    }

    if foundDebuff != nil {
      newStack := EffectStack{
        TotalDmg: effect.BaseDmg,
        PlusCrit: effect.PlusCrit,
        PlusHit: effect.PlusHit,
        PlusDmg: effect.PlusDmg,
      }
      
      if len(foundDebuff.Stacks) < foundDebuff.MaxStacks {
        if LOG_ADD_DEBUFF_STACK {
          log.Printf("???Adding new stack: %f (%s)\n", newStack.TotalDmg, foundDebuff.Name)
        }
        foundDebuff.Stacks = append(foundDebuff.Stacks, newStack)
      } else if effect.ApplyMask&(DEBUFF_MY_OVERRIDE_SMALLEST|DEBUFF_OTHER_OVERRIDE_SMALLEST) != 0 {
        smallestStack := &newStack
        for i, stack := range foundDebuff.Stacks {
          if stack.TotalDmg < smallestStack.TotalDmg {
            smallestStack = &foundDebuff.Stacks[i]
          }
        }
        
        if smallestStack.TotalDmg != newStack.TotalDmg {
          if LOG_OVERRIDE_DEBUFF_STACK {
            log.Printf("!!!Replacing smallest: %f with new %f (%s)\n", smallestStack.TotalDmg, newStack.TotalDmg, foundDebuff.Name)
          }
          *smallestStack = newStack
        }
      }

      if effect.ApplyMask&DEBUFF_RESET_DURATION != 0 {
        foundDebuff.TimeRefreshed = time.Now()
      }
      
      return
    }
  }

  newDebuff := Debuff{
    Effect: effect,
    TimeApplied: time.Now(),
    TimeRefreshed: time.Now(),
    ApplierName: mage.Name,
    Stacks: []EffectStack{
      EffectStack{
        TotalDmg: effect.BaseDmg,
        PlusCrit: effect.PlusCrit,
        PlusHit: effect.PlusHit,
        PlusDmg: effect.PlusDmg,
      },
    },
  }

  if LOG_ADD_DEBUFF_INSTANCE {
    log.Printf("$$$new debuff: %f (%s)\n", newDebuff.Stacks[0].TotalDmg, newDebuff.Name)
  }
  mob.Debuffs = append(mob.Debuffs, newDebuff)
}

func (mob *Monster) Hit(mage Mage, spell Spell) AttackResult {
  var attackResult AttackResult
  var damage float32

  hitRoll := rand.Float32() * 100
  if hitRoll < 100 - mob.BaseMiss + spell.Hit + mage.Hit {
    attackResult = HIT

    damage = spell.BaseDmgMin + rand.Float32() * (spell.BaseDmgMax - spell.BaseDmgMin)
    damage += float32(mage.SP) * spell.SPScaling
    damage *= mage.DmgScale + mob.GetPlusDmg()/100

    critRoll := rand.Float32() * 100
    if critRoll < spell.Crit + mage.Crit {
      attackResult = CRIT
      damage *= 1.5
    }
  } else {
    attackResult = MISS
  }

  for _, effect := range spell.Effects {
    effect.BaseDmg += damage * effect.SpellDmg
    mob.ApplyDebuff(mage, effect, attackResult)
  }

  mob.DmgTaken += float64(damage)

  if LOG_INITIAL_DMG {
    log.Printf("%f %s [%s]\n", damage, spell.Name, mage.Name)
  }
  
  return attackResult
}

func (mob *Monster) Play(stop chan int) {
  for {
    select {
    case <-stop:
      return
    default:
      for i := len(mob.Debuffs)-1; i >= 0; i-- {
        debuff := mob.Debuffs[i]
      
        if debuff.Interval.Nanoseconds() != 0 {
          lastTime := debuff.TimeTicked
          if lastTime == (time.Time{}) {
            lastTime = debuff.TimeApplied
          }
      
          if lastTime.Add(ScaleDuration(debuff.Interval)).Before(time.Now()) {
            var dotDmg float64
            for _, stack := range debuff.Stacks {
              dotDmg += debuff.Interval.Seconds() / debuff.Duration.Seconds() * float64(stack.TotalDmg)
            }
            
            dotDmg = math.Floor(dotDmg)
            dotDmg *= float64(1 + mob.GetPlusDmg()/100)
            dotDmg = math.Floor(dotDmg)
            
            if LOG_DOT_DMG {
              log.Printf("%f %s\n", dotDmg, debuff.Name)
            }
            mob.DmgTaken += dotDmg
      
            mob.Debuffs[i].TimeTicked = time.Now()
          }
        }
      
        if debuff.TimeRefreshed.Add(ScaleDuration(debuff.Duration)).Add(ScaleDuration(50 * time.Millisecond)).Before(time.Now()) {
          if LOG_DEBUFF_DROPPED {
            log.Printf("-----Debuff '%s' DROPPED\n", debuff.Name)
          }
          mob.Debuffs = append(mob.Debuffs[:i], mob.Debuffs[i+1:]...)
        }
      }
    }
  }
}

type Mage struct {
  Name string
  SP int32
  Crit float32
  Hit float32
  DmgScale float32 // +x% increase (e.g. Fire Power talent)
  StartDelay time.Duration // delay before starting to cast 1st spell
  Strategy func(m Mage, start time.Time)
}

type Monster struct {
  Name string
  DmgTaken float64
  Debuffs []Debuff
  BaseMiss float32 // chance to miss this monster
}

type Debuff struct {
  Effect

  Stacks []EffectStack
  TimeTicked time.Time // time of last dot tick
  TimeApplied time.Time
  TimeRefreshed time.Time
  ApplierName string
}

type DebuffApply uint32

const (
  DEBUFF_MY_NEW_INSTANCE DebuffApply = 1 << iota // my debuff exists - create new instance
  DEBUFF_OTHER_NEW_INSTANCE // someones debuff exists - create new instance
  DEBUFF_MY_NEW_STACK // my debuff exists - increment stacks (below max)
  DEBUFF_OTHER_NEW_STACK // someones debuff exists - increment stacks (below max)
  DEBUFF_MY_OVERRIDE_SMALLEST // my debuff exists - override if dmg is higher
  DEBUFF_OTHER_OVERRIDE_SMALLEST // others debuff exists - override if dmg is higher
  DEBUFF_RESET_DURATION // reset debuff duration
)

type AttackResult int8

const (
  _ AttackResult = 0
  MISS AttackResult = 1
  HIT AttackResult = 2
  CRIT AttackResult = 3
)

type Spell struct {
  Name string
  CastTime time.Duration
  BaseDmgMin float32
  BaseDmgMax float32
  SPScaling float32
  Crit float32
  Hit float32
  Effects []Effect
}

type Effect struct {
  Name string
  MaxStacks int
  BaseDmg float32
  SpellDmg float32 // +x% of dmg from initial hit as fraction
  SPScaling float32
  PlusCrit float32
  PlusHit float32
  PlusDmg float32 // +x% dmg increase
  Interval time.Duration
  Duration time.Duration
  ApplyMask DebuffApply
  Condition AttackResult
}

type EffectStack struct {
  TotalDmg float32 // total dmg throughout duration
  PlusCrit float32
  PlusHit float32
  PlusDmg float32
}

var Ignite = Effect {
  Name: "Ignite",
  MaxStacks: 5,
  SpellDmg: 0.4,
  SPScaling: 0,
  Interval: 2 * time.Second,
  Duration: 4 * time.Second,
  Condition: CRIT,
  ApplyMask: (DEBUFF_MY_NEW_STACK|DEBUFF_OTHER_NEW_STACK|DEBUFF_MY_OVERRIDE_SMALLEST|DEBUFF_OTHER_OVERRIDE_SMALLEST|DEBUFF_RESET_DURATION),
}

var Scorch = Spell {
  Name: "Scorch",
  CastTime: 1500 * time.Millisecond,
  BaseDmgMin: 233,
  BaseDmgMax: 276,
  SPScaling: 0.4285,
  Crit: 4,
  Effects: []Effect{
    Ignite,
    Effect {
      Name: "Fire Vulnerability",
      MaxStacks: 5,
      Duration: 30 * time.Second,
      PlusDmg: 3,
      Condition: HIT,
      ApplyMask: (DEBUFF_MY_NEW_STACK|DEBUFF_OTHER_NEW_STACK|DEBUFF_RESET_DURATION),
    },
  },
}

var Fireball = Spell {
  Name: "Fireball",
  CastTime: 3 * time.Second,
  BaseDmgMin: 561,
  BaseDmgMax: 716,
  SPScaling: 1,
  Effects: []Effect{
    Ignite,
    Effect {
      Name: "Fireball (dot)",
      BaseDmg: 78, // 72 base + increase from Fire Power
      SPScaling: 0,
      Interval: 2 * time.Second,
      Duration: 8 * time.Second,
      Condition: HIT,
      ApplyMask: (DEBUFF_MY_NEW_STACK|DEBUFF_RESET_DURATION),
    },
  },
}

func ScaleDuration(duration time.Duration) time.Duration{
  return time.Duration(float32(duration) / TimeScale)
}

var wg sync.WaitGroup
var lock sync.Mutex

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
    return fmt.Print(time.Now().UTC().Format("15:04:05.999Z ") + string(bytes))
}

func RunSimulation() float64 {
  stop := make(chan int)

  go TARGET.Play(stop)
  for _, mage := range Mages {
    wg.Add(1)
    go func(mage Mage) {
      time.Sleep(ScaleDuration(mage.StartDelay))
      mage.Play()
    }(mage)
  }

  wg.Wait()
  
  stop <- 0
  
  return TARGET.DmgTaken
}

func main() {
  rand.Seed(time.Now().UnixNano())
  log.SetFlags(0)
  log.SetOutput(new(logWriter))
  
  var totalDpsPerMage []float32
  for i := 1; i <= SimulationRunCount; i++ {
    TARGET.DmgTaken = 0
    TARGET.Debuffs = []Debuff{}
    
    log.Printf("   Starting simulation #%d. Run time: %ds...\n", i, int(SimulationDuration.Seconds()))
    
    RunSimulation()
    
    log.Println()
    
    //log.Printf("Duration: %ds\n", int(SimulationDuration.Seconds()))
    
    //log.Printf("Damage done: %f\n", TARGET.DmgTaken)
    //log.Printf("Damage done (per mage): %f\n", TARGET.DmgTaken / float64(len(Mages)))
   
    //log.Printf("DPS: %f\n", TARGET.DmgTaken / SimulationDuration.Seconds())
    dpsPerMage := TARGET.DmgTaken / float64(len(Mages)) / SimulationDuration.Seconds()
    totalDpsPerMage = append(totalDpsPerMage, float32(dpsPerMage))
    log.Printf("DPS (per mage): %f\n", dpsPerMage)
    
    time.Sleep(100 * time.Millisecond)
  }
  
  log.Println()
  log.Printf("Total DPS per mage: %v\n", totalDpsPerMage)
  
  var total float32 = 0
  for _, dps := range totalDpsPerMage {
    total += dps
  }
  log.Printf("AVG Total DPS per mage: %f\n", total/float32(len(totalDpsPerMage)))
  
  log.Println("   FIN")
}
