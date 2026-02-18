# GitHub Repository İyileştirme Rehberi

Bu rehber, GitHub repository'nizi daha görünür ve profesyonel hale getirmek için yapmanız gereken iyileştirmeleri içerir.

## ✅ Otomatik Olarak Hazırlanan Dosyalar

Aşağıdaki dosyalar oluşturuldu ve commit edilmeye hazır:

- ✅ `CHANGELOG.md` - Değişiklik geçmişi
- ✅ `RELEASE_NOTES_v0.1.0.md` - İlk release notları
- ✅ `.github/RELEASE_TEMPLATE.md` - Gelecek release'ler için şablon

## 📋 GitHub'da Yapılacaklar (Manuel)

### 1. Repository Description ve About Bölümü

1. Repository sayfasına gidin: https://github.com/mos1907/GO-NMOS
2. Sağ üstteki **⚙️ Settings** butonuna tıklayın
3. **General** sekmesinde aşağı kaydırın
4. **About** bölümünde:
   - **Description**: 
     ```
     Production-oriented NMOS management stack using Go + Svelte. IS-04/IS-05 discovery, patch panel, collision detection, and automation.
     ```
   - **Website** (opsiyonel): Eğer bir demo sitesi varsa
   - **Topics** (etiketler) - Şunları ekleyin:
     ```
     nmos
     go
     svelte
     broadcast
     media-production
     docker
     postgresql
     mqtt
     is-04
     is-05
     tailwindcss
     ```

### 2. İlk Release Oluşturma

1. Repository sayfasında **"Releases"** sekmesine gidin
2. **"Create a new release"** butonuna tıklayın
3. **Choose a tag**: `v0.1.0` yazın (yeni tag oluştur)
4. **Release title**: `v0.1.0 - Initial Release`
5. **Description**: `RELEASE_NOTES_v0.1.0.md` dosyasının içeriğini kopyalayıp yapıştırın
6. **"Publish release"** butonuna tıklayın

### 3. Repository Topics (Etiketler) Ekleme

**Yöntem 1: Settings'den**
1. Settings → General → About bölümünde Topics alanına ekleyin

**Yöntem 2: Repository Ana Sayfasından**
1. Repository ana sayfasında, sağ üstte **"Add topics"** butonuna tıklayın
2. Şu etiketleri ekleyin:
   - `nmos`
   - `go`
   - `svelte`
   - `broadcast`
   - `media-production`
   - `docker`
   - `postgresql`
   - `mqtt`
   - `is-04`
   - `is-05`
   - `tailwindcss`

### 4. README Badge'leri (Opsiyonel)

README.md dosyasının başına badge'ler eklenmiş. Eğer GitHub Actions çalışıyorsa, workflow badge'leri de eklenebilir.

### 5. Community Standards

GitHub otomatik olarak şunları kontrol eder:
- ✅ README.md (var)
- ✅ LICENSE (var)
- ✅ CONTRIBUTING.md (var)
- ✅ Issue templates (var)
- ✅ Pull request template (var)
- ⚠️ Code of Conduct (opsiyonel - eklenebilir)

**Code of Conduct eklemek isterseniz:**
1. `.github/CODE_OF_CONDUCT.md` dosyası oluşturun
2. [Contributor Covenant](https://www.contributor-covenant.org/) kullanabilirsiniz

### 6. Branch Protection (Önerilir)

1. Settings → Branches
2. **"Add branch protection rule"**
3. Branch name pattern: `main`
4. Ayarlar:
   - ✅ Require a pull request before merging
   - ✅ Require status checks to pass before merging
     - `pr-check` workflow'unu seçin
   - ✅ Require conversation resolution before merging
   - ✅ Include administrators (opsiyonel)

### 7. GitHub Actions Workflow'ları

`.github/workflows/pr-check.yml` dosyası hazır. İlk PR'da otomatik çalışacak.

## 🚀 Hızlı Komutlar

### Dosyaları Commit ve Push Etme

```bash
cd /Users/muratdemirci/GO-NMOS

# Yeni dosyaları ekle
git add CHANGELOG.md RELEASE_NOTES_v0.1.0.md .github/RELEASE_TEMPLATE.md

# Commit et
git commit -m "docs: add changelog and release notes for v0.1.0"

# Push et
git push origin main
```

### Release Tag Oluşturma (Opsiyonel - GitHub'dan da yapılabilir)

```bash
# Tag oluştur
git tag -a v0.1.0 -m "Initial release v0.1.0"

# Tag'i push et
git push origin v0.1.0
```

## 📊 İyileştirme Kontrol Listesi

- [ ] Repository description eklendi
- [ ] Topics/etiketler eklendi
- [ ] İlk release (v0.1.0) oluşturuldu
- [ ] CHANGELOG.md commit edildi
- [ ] RELEASE_NOTES_v0.1.0.md commit edildi
- [ ] Branch protection kuralları ayarlandı (opsiyonel)
- [ ] Code of Conduct eklendi (opsiyonel)

## 🎯 Sonuç

Bu iyileştirmeleri tamamladıktan sonra:
- ✅ Repository daha profesyonel görünecek
- ✅ Arama motorlarında daha iyi bulunacak
- ✅ Katkıda bulunmak isteyenler için daha açık olacak
- ✅ GitHub'ın "Community standards" badge'ini alacak

## 📝 Notlar

- Release oluşturduktan sonra GitHub otomatik olarak bir ZIP dosyası oluşturur
- Topics ekledikten sonra GitHub'ın keşif özelliklerinde görünürsünüz
- Description ve topics, repository'nin GitHub aramasında bulunabilirliğini artırır

Başarılar! 🚀
