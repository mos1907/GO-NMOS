# GitHub'da Açık Kaynak Proje Oluşturma Rehberi

Bu rehber, go-NMOS projenizi GitHub'da herkesin katkıda bulunabileceği açık kaynak bir proje haline getirmenize yardımcı olur.

## 1. GitHub Repository Oluşturma

1. GitHub'a giriş yapın: https://github.com
2. Sağ üstteki **"+"** butonuna tıklayın → **"New repository"**
3. Repository bilgilerini doldurun:
   - **Repository name**: `GO-NMOS` (veya istediğiniz isim)
   - **Description**: "Production-oriented NMOS management stack using Go + Svelte"
   - **Visibility**: ✅ **Public** (açık kaynak için)
   - **Initialize repository**: ❌ Boş bırakın (zaten kod var)
4. **"Create repository"** butonuna tıklayın

## 2. Projeyi GitHub'a Push Etme

Eğer projeniz henüz git repository değilse:

```bash
cd /Users/muratdemirci/GO-NMOS

# Git repository başlat
git init

# Tüm dosyaları ekle
git add .

# İlk commit
git commit -m "Initial commit: go-NMOS project"

# GitHub repository'nizi remote olarak ekle
git remote add origin https://github.com/mos1907/GO-NMOS.git

# Ana branch'i main olarak ayarla
git branch -M main

# GitHub'a push et
git push -u origin main
```

Eğer zaten git repository ise:

```bash
# GitHub repository'nizi remote olarak ekle
git remote add origin https://github.com/mos1907/GO-NMOS.git

# GitHub'a push et
git push -u origin main
```

## 3. GitHub Repository Ayarları

### 3.1. Repository Ayarları

1. Repository sayfasında **"Settings"** sekmesine gidin
2. **"General"** bölümünde:
   - **Features**:
     - ✅ Issues (açık bırakın)
     - ✅ Discussions (isteğe bağlı)
     - ✅ Projects (isteğe bağlı)
     - ✅ Wiki (isteğe bağlı)
   - **Pull Requests**:
     - ✅ "Allow merge commits"
     - ✅ "Allow squash merging"
     - ✅ "Allow rebase merging"

### 3.2. Branch Protection (Önerilir)

1. **Settings** → **Branches**
2. **"Add branch protection rule"**
3. **Branch name pattern**: `main` (veya `master`)
4. Ayarlar:
   - ✅ "Require a pull request before merging"
   - ✅ "Require status checks to pass before merging"
     - CI workflow'unuzu seçin (varsa)
   - ✅ "Require conversation resolution before merging"
   - ✅ "Include administrators" (opsiyonel)

### 3.3. Topics (Etiketler) Ekleme

Repository ana sayfasında **"Add topics"** butonuna tıklayın ve şunları ekleyin:
- `nmos`
- `go`
- `svelte`
- `broadcast`
- `media-production`
- `docker`
- `postgresql`

## 4. README.md Güncelleme

README.md dosyasında GitHub repository URL'lerini güncelleyin:

```markdown
# README.md içinde şu satırları bulun ve güncelleyin:

[open an issue](https://github.com/mos1907/GO-NMOS/issues)
```

✅ GitHub kullanıcı adı `mos1907` olarak ayarlandı.

## 5. GitHub Actions Workflow Kontrolü

`.github/workflows/` klasöründe CI workflow'unuzun olduğundan emin olun. Eğer yoksa, `pr-check.yml` dosyası oluşturuldu.

## 6. İlk Issue ve PR Oluşturma (Test)

1. **Test Issue Oluşturma**:
   - Repository'de **"Issues"** sekmesine gidin
   - **"New issue"** butonuna tıklayın
   - Template'lerden birini seçin (Bug Report veya Feature Request)
   - Test için bir issue oluşturun

2. **Test PR Oluşturma**:
   - Bir branch oluşturun: `git checkout -b test/readme-update`
   - README'de küçük bir değişiklik yapın
   - Commit edin: `git commit -m "docs: update README"`
   - Push edin: `git push origin test/readme-update`
   - GitHub'da **"Compare & pull request"** butonuna tıklayın
   - PR template'inin çalıştığını kontrol edin

## 7. Community Standards (Önerilir)

GitHub'da **"Community standards"** badge'i almak için:

1. **Settings** → **General** → en alta scroll edin
2. **"Community standards"** bölümünde eksikleri tamamlayın:
   - ✅ README.md (var)
   - ✅ LICENSE (var)
   - ✅ CONTRIBUTING.md (var)
   - ✅ Code of Conduct (opsiyonel, eklenebilir)
   - ✅ Issue templates (var)
   - ✅ Pull request template (var)

### Code of Conduct Eklemek (Opsiyonel)

```bash
# GitHub'ın standart Contributor Covenant Code of Conduct'unu kullanabilirsiniz
# https://www.contributor-covenant.org/
```

## 8. GitHub Pages (Opsiyonel - Dokümantasyon için)

Eğer dokümantasyon sitesi oluşturmak isterseniz:

1. **Settings** → **Pages**
2. **Source**: `main` branch, `/docs` folder
3. Veya GitHub Actions ile otomatik deploy

## 9. Release Oluşturma

İlk release'i oluşturun:

1. **"Releases"** → **"Create a new release"**
2. **Tag version**: `v0.1.0`
3. **Release title**: `v0.1.0 - Initial Release`
4. **Description**: İlk release notları
5. **"Publish release"**

## 10. Katkıda Bulunanları Teşvik Etme

- README'de "Contributors welcome" mesajı ekleyin ✅
- İyi first issues etiketleyin (yeni başlayanlar için)
- Pull request'lere hızlı geri bildirim verin
- Code review yapın
- Teşekkür edin! 🙏

## Hızlı Kontrol Listesi

- [ ] GitHub repository oluşturuldu ve Public
- [ ] Kod GitHub'a push edildi
- [ ] README.md güncellendi (GitHub URL'leri)
- [ ] LICENSE dosyası eklendi
- [ ] CONTRIBUTING.md eklendi
- [ ] Issue templates eklendi
- [ ] Pull request template eklendi
- [ ] .gitignore dosyası kontrol edildi
- [ ] GitHub Actions workflow çalışıyor
- [ ] İlk release oluşturuldu
- [ ] Topics/etiketler eklendi

## Sonraki Adımlar

1. **Projeyi tanıtın**: Reddit, Hacker News, Twitter/X, LinkedIn
2. **Dokümantasyonu geliştirin**: API docs, örnekler, tutorial'lar
3. **Community oluşturun**: Discussions açın, sorulara cevap verin
4. **Düzenli güncellemeler**: Release notes, changelog tutun

## Yardımcı Linkler

- [GitHub Open Source Guide](https://opensource.guide/)
- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)

Başarılar! 🚀
