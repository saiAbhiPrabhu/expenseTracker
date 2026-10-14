# PERFECT_FUNCTION.md — Function Design Rules
 
Apply to every function you write or edit. Goal: functions that support **local reasoning** (understandable from the body alone) and **testability**. Examples are C++-flavored; translate the idea to the language at hand.
 
## The 5 Tenets
 
1. **Be honest.** The signature is the complete contract. Read/write nothing that isn't reachable through the parameters. No hidden globals, clock, RNG, env, singletons, or I/O. (Mutating an argument you were explicitly handed is still honest.)
2. **Dishonesty goes at the top.** Dishonesty is infectious: an honest function can't call a dishonest one. Build logic from honest functions at the leaves; do I/O, time, randomness, and globals in a thin outer shell (`main`, handlers, framework hooks) and pass results in. If materializing output is too costly, inject a callback (`for_each_x(input, fn)`).
3. **Serve the caller.** The signature is read first by a human. No mystery `bool`s or long same-typed arg lists (use an options struct or strong types). Accept the most general input you actually need (`span`/iterable, not `vector`), and only the fields you use (not the whole object). Never demand stronger guarantees than the body needs.
4. **Make preconditions unrepresentable.** If a parameter has an invariant, encode it in a type (`Normalized<Vec3>`, a lock-proof token, `NonEmpty<T>`) that can only be built validly. Don't re-validate, assert, or return errors for something the type can guarantee. Push the failure to the caller, who knows more.
5. **One level of abstraction per body.** Every line sits at the same level; stack bricks (existing functions), never zoom in and out. Section comments, raw loops with nontrivial bodies, and hand-rolled algorithms (binary search, case-folding) are zoom-ins: extract them or use the library. Extract at **n = 1**, not n = 3. Write the call you wish existed, then implement it. Framework hooks (`update`, `main`) stay tiny and delegate immediately.
## Examples
 
### 1. Dishonest accessor → honest (T1)
```cpp
// AVOID: name promises "all required"; secretly filters by global load state
vector<Asset> GetRequiredAssets() { return filter(all, [](a){ return !AssetMgr::Get().IsLoaded(a); }); }
 
// STRIVE: honest accessor; caller filters explicitly
vector<Asset> GetRequiredAssets(const Level&);
auto toLoad = filter(GetRequiredAssets(lvl), [&](a){ return !mgr.IsLoaded(a); });
```
 
### 2. Hidden global RNG → injected (T1)
```cpp
// AVOID: untestable, irreproducible, action-at-a-distance
void World::Populate(int n) { for (...) p.pos = GlobalRng().Next(); }
 
// STRIVE: behavior fully determined by args; seed (time/OS/fixed) chosen at the top
void World::Populate(int n, Rng& rng);
```
 
### 3. Logic tangled with I/O → core + shell (T2)
```cpp
// AVOID: permutation logic and printing interleaved; only testable by running the CLI
int main() { string s; cin >> s; do { cout << s; } while (next_permutation(...)); }
 
// STRIVE: honest core, dishonest skin
void ForEachPermutation(string s, function<void(const string&)> visit);   // or return vector
int main() { string s; cin >> s; ForEachPermutation(s, [](auto& p){ cout << p << '\n'; }); }
```
 
### 4. Mystery args → options struct (T3)
```cpp
// AVOID
Render(scene, 800, 600, true, false, 2);
 
// STRIVE
Render(scene, {.width = 800, .height = 600, .vsync = true, .msaa = 2});
```
 
### 5. Over-constrained params → general params (T3)
```cpp
// AVOID: forces my storage; takes whole wallet when only the card is needed
int Sum(const std::vector<int>&);        void Charge(const Customer&);
 
// STRIVE: ask for exactly what you use
int Sum(std::span<const int>);           void Charge(const PaymentCard&);
```
 
### 6. Unchecked precondition → validating type (T4)
```cpp
// AVOID: only correct if `n` has length 1; callers silently get NaNs, or you add a runtime check
Vec3 Bounce(Vec3 v, Vec3 n);
 
// STRIVE: invariant lives in the type; Bounce has no error path
Vec3 Bounce(Vec3 v, NormalizedVec3 n);     // NormalizedVec3 converts implicitly to Vec3
```
 
### 7. Hidden ordering dependency → proof token (T4)
```cpp
// AVOID: "// caller must hold mutex_"
void B();
 
// STRIVE: cannot call B without having locked
void B(const std::lock_guard<std::mutex>& proof);
```
 
### 8. Mixed abstraction levels → bricks (T5)
```cpp
// AVOID: lowercase loop + hand-rolled binary search + type check, with section comments
bool IsAssetOfType(string name, Type t) {
  // lowercase
  for (char& c : name) c = tolower(c);
  // binary search
  size_t lo = 0, hi = assets.size(); while (lo < hi) { /* ... */ }
  if (not found) exit(1);                  // also: never kill the process in a library fn
  return assets[lo].type == t;
}
 
// STRIVE: every line at the same level; details live in tested bricks
bool IsAssetOfType(const string& name, Type t) {
  auto asset = assets.Find(name);          // CaseInsensitiveMap<Asset> owns lowercase + search
  return asset && asset->type == t;
}
```
 
### 9. Fat framework hook → thin glue (T5, T2)
```cpp
// AVOID: 200 lines of physics, input, and drawing inside the callback
void Update(float dt) { /* everything */ }
 
// STRIVE: hook only bridges framework to your code
void Update(float dt) { world = Step(std::move(world), ReadInput(), dt); Draw(world); }
```
 
## Pre-Commit Checklist
- Can I test this with no global setup? If not, which tenet (1/2) is violated?
- Does the signature alone tell a caller everything it reads, writes, and requires?
- Does any line need a "zoom in" comment, a raw loop, or a hand-rolled algorithm? Extract it.
- Is any precondition enforced by a comment or `assert` that a type could enforce?
- Is dishonest work (I/O, time, RNG, globals) as close to the root of the call tree as possible?
*Source: "How to write the perfect function" talk (honesty / call-tree dishonesty / empathetic signatures / one level of abstraction).*
