using System;
using System.Data;
using System.Collections.Generic;
using System.Linq;
using System.Web;
using System.Text;
using System.Web.Mail;
using System.Web.UI.WebControls;

public class Ayarlar
{
    public static DataRow AyarlarTabloRS;
    //------------------------------------------------------------------------
    public static void BaglantiTablosu()
    {
        if (AyarlarTabloRS == null)
        { AyarlarTabloRS = Baglanti.GetDataRow("select * from Tbl_SiteAyarlari"); }
    }
    //------------------------------------------------------------------------
    public static int UrlLinkTur
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["UrlLinkTur"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string FirmaAdi
    { get { BaglantiTablosu(); return AyarlarTabloRS["FirmaAdi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string Email
    { get { BaglantiTablosu(); return AyarlarTabloRS["Email"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string WebAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["WebAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string Siteismi
    { get { BaglantiTablosu(); return AyarlarTabloRS["Siteismi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string GoogleBaslik
    { get { BaglantiTablosu(); return AyarlarTabloRS["GoogleBaslik"].ToString(); } }
    //-----------------------------------------------------------------------
    public static bool RobotsDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["RobotsDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string Description
    { get { BaglantiTablosu(); return AyarlarTabloRS["Description"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string Keywords
    { get { BaglantiTablosu(); return AyarlarTabloRS["Keywords"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string GoogleAnalitik
    { get { BaglantiTablosu(); return AyarlarTabloRS["GoogleAnalitik"].ToString().Replace("´", "'"); } }
    //-----------------------------------------------------------------------
    public static string GoogleDogrulamaKodu
    { get { BaglantiTablosu(); return AyarlarTabloRS["GoogleDogrulamaKodu"].ToString().Replace("´", "'"); } }
    //-----------------------------------------------------------------------
    public static string YandexAnalitik
    { get { BaglantiTablosu(); return AyarlarTabloRS["YandexAnalitik"].ToString().Replace("´", "'"); } }
    //-----------------------------------------------------------------------
    public static string YandexDogrulamaKodu
    { get { BaglantiTablosu(); return AyarlarTabloRS["YandexDogrulamaKodu"].ToString().Replace("´", "'"); } }
    //-----------------------------------------------------------------------
    public static string GeoTags
    { get { BaglantiTablosu(); return AyarlarTabloRS["GeoTags"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string EkMetaTaglar
    { get { BaglantiTablosu(); return AyarlarTabloRS["EkMetaTaglar"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string GenelMailAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["GenelMailAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string PopupGenislik
    { get { BaglantiTablosu(); return AyarlarTabloRS["PopupGenislik"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string PopupYukseklik
    { get { BaglantiTablosu(); return AyarlarTabloRS["PopupYukseklik"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string PopupTur
    { get { BaglantiTablosu(); return AyarlarTabloRS["PopupTur"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string PopupSuresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["PopupSuresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static bool PopupDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["PopupDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string PopupResim
    { get { BaglantiTablosu(); return AyarlarTabloRS["PopupResim"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string Popupicerik
    { get { BaglantiTablosu(); return AyarlarTabloRS["Popupicerik"].ToString(); } }
    //----------------------------------------------------------------------
    public static string PopupUrl
    { get { BaglantiTablosu(); return AyarlarTabloRS["PopupUrl"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string PopupAcilisSekli
    { get { BaglantiTablosu(); return AyarlarTabloRS["PopupAcilisSekli"].ToString(); } }
    //---------------------------------------------------------------------
    public static bool LogRaporDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["LogRaporDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool VeriRaporDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["VeriRaporDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool PingDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["PingDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string FirmaSiteLogo
    { get { BaglantiTablosu(); return AyarlarTabloRS["FirmaSiteLogo"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string FirmaLogo
    { get { BaglantiTablosu(); return AyarlarTabloRS["FirmaLogo"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string Firmaicon
    { get { BaglantiTablosu(); return AyarlarTabloRS["Firmaicon"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string FacebookAppID
    { get { BaglantiTablosu(); return AyarlarTabloRS["FacebookAppID"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string FacebookAdmin
    { get { BaglantiTablosu(); return AyarlarTabloRS["FacebookAdmin"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string FacebookAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["FacebookAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string TwiterNickName
    { get { BaglantiTablosu(); return AyarlarTabloRS["TwiterNickName"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string TwiterAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["TwiterAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string YouTubeAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["YouTubeAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string InstagramAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["InstagramAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string VimeoAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["VimeoAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string FriendFeed
    { get { BaglantiTablosu(); return AyarlarTabloRS["FriendFeed"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string LinkedinAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["LinkedinAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string PinterestAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["PinterestAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string SkypeAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["SkypeAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string GooglePlus
    { get { BaglantiTablosu(); return AyarlarTabloRS["GooglePlus"].ToString(); } }
    //-----------------------------------------------------------------------
    public static bool WatermakPanelDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["WatermakPanelDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool WatermakDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["WatermakDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string WatermakResim
    { get { BaglantiTablosu(); return AyarlarTabloRS["WatermakResim"].ToString(); } }
    //-----------------------------------------------------------------------
    public static bool ResimLogoDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["ResimLogoDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string ResimLogo
    { get { BaglantiTablosu(); return AyarlarTabloRS["ResimLogo"].ToString(); } }
    //-----------------------------------------------------------------------
    public static bool ArkaPlanResimDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["ArkaPlanResimDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string ArkaPlanResim
    { get { BaglantiTablosu(); return AyarlarTabloRS["ArkaPlanResim"].ToString(); } }
    //-----------------------------------------------------------------------
    public static bool SiteAltiicerikDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["SiteAltiicerikDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool SiteAltiLinkDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["SiteAltiLinkDurum"].ToString()); } }
    //----------------------------------------------------------------------
    public static bool SiteAltiAltsayfaDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["SiteAltiAltsayfaDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string SiteAltiicerik
    { get { BaglantiTablosu(); return AyarlarTabloRS["SiteAltiicerik"].ToString(); } }
    //-----------------------------------------------------------------------
    public static bool SagClickDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["SagClickDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int KucukResimGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["KucukResimGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int OrtaResimGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["OrtaResimGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int BuyukResimGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["BuyukResimGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int CokluKucukResimGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["CokluKucukResimGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int CokluOrtaResimGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["CokluOrtaResimGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int CokluBuyukResimGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["CokluBuyukResimGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int MenuKucukResimGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["MenuKucukResimGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int MenuOrtaResimGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["MenuOrtaResimGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int MenuBuyukResimGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["MenuBuyukResimGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int KucukResimYukseklik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["KucukResimYukseklik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int OrtaResimYukseklik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["OrtaResimYukseklik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int BuyukResimYukseklik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["BuyukResimYukseklik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int CokluKucukResimYukseklik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["CokluKucukResimYukseklik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int CokluOrtaResimYukseklik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["CokluOrtaResimYukseklik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int CokluBuyukResimYukseklik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["CokluBuyukResimYukseklik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int MenuKucukResimYukseklik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["MenuKucukResimYukseklik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int MenuOrtaResimYukseklik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["MenuOrtaResimYukseklik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int MenuBuyukResimYukseklik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["MenuBuyukResimYukseklik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int BannerKucukResimGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["BannerKucukResimGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int BannerOrtaResimGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["BannerOrtaResimGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int BannerBuyukResimGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["BannerBuyukResimGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int BannerKucukResimYukseklik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["BannerKucukResimYukseklik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int BannerOrtaResimYukseklik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["BannerOrtaResimYukseklik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int BannerBuyukResimYukseklik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["BannerBuyukResimYukseklik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string MobilLogo
    { get { BaglantiTablosu(); return AyarlarTabloRS["MobilLogo"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string Yardim
    { get { BaglantiTablosu(); return AyarlarTabloRS["Yardim"].ToString(); } }
    //-----------------------------------------------------------------------
    public static bool ResimYuklemeTur
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["ResimYuklemeTur"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool MenuAltKategoriDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["MenuAltKategoriDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool MenuKategoriDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["MenuKategoriDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool TekliResimKaliteDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["TekliResimKaliteDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool KucukResimKaliteDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["KucukResimKaliteDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool OrtaResimKaliteDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["OrtaResimKaliteDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool BuyukResimKaliteDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["BuyukResimKaliteDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool ResimSilmeDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["ResimSilmeDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool TopluKategoriKayitSilmeDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["TopluKategoriKayitSilmeDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool TopluKayitSilmeDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["TopluKayitSilmeDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool TopluResimYuklemeDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["TopluResimYuklemeDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool TopluDosyaYuklemeDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["TopluDosyaYuklemeDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool TopluKatalogYuklemeDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["TopluKatalogYuklemeDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool ResimBaslikAdiDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["ResimBaslikAdiDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool AcilirMenuDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["AcilirMenuDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool MenuYeniKayitDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["MenuYeniKayitDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool MenuKategoriYetkiDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["MenuKategoriYetkiDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool ListeKategoriDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["ListeKategoriDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool FormExcelDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["FormExcelDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool CropDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["CropDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool CropKaydetveDevam
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["CropKaydetveDevam"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool CropKapsamliDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["CropKapsamliDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool CropResimKopyalama
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["CropResimKopyalama"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool CropBuyukUygunDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["CropBuyukUygunDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool CropEkGenislikDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["CropEkGenislikDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int CropEkGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["CropEkGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int CropBuyukGenislikResimKati
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["CropBuyukGenislikResimKati"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool CropBuyukGenislikResimKatiDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["CropBuyukGenislikResimKatiDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string GoogleAnalitikEmail
    { get { BaglantiTablosu(); return AyarlarTabloRS["GoogleAnalitikEmail"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string GoogleAnalitikSifre
    { get { BaglantiTablosu(); return AyarlarTabloRS["GoogleAnalitikSifre"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string GoogleAnalitikProfilID
    { get { BaglantiTablosu(); return AyarlarTabloRS["GoogleAnalitikProfilID"].ToString(); } }
    //-----------------------------------------------------------------------
    public static bool GenelSilDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["GenelSilDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool MenuKategoriSilDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["MenuKategoriSilDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool SayfaTuruDuzenle
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["SayfaTuruDuzenle"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool SayfaTuruYeniKayit
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["SayfaTuruYeniKayit"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string TelefonNumarasi
    { get { BaglantiTablosu(); return AyarlarTabloRS["TelefonNumarasi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static bool SiteMapDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["SiteMapDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool RssDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["RssDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool SeoYonlendirmeDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["SeoYonlendirmeDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool CacheDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["CacheDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int MenuListemeSayisi
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["MenuListemeSayisi"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool iliskiliKayitDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["iliskiliKayitDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string Aciklama
    { get { BaglantiTablosu(); return AyarlarTabloRS["Aciklama"].ToString(); } }
    //-----------------------------------------------------------------------
    public static bool GenelHataDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["GenelHataDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string GenelHataEmailAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["GenelHataEmailAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string CanliDestekKodu
    { get { BaglantiTablosu(); return AyarlarTabloRS["CanliDestekKodu"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string FooterAlani
    { get { BaglantiTablosu(); return AyarlarTabloRS["FooterAlani"].ToString().Replace("´", "'"); } }
    //-----------------------------------------------------------------------
    public static bool TumGenelKategoriGlobalOlsun
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["TumGenelKategoriGlobalOlsun"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool SeoUrlKontrolDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["SeoUrlKontrolDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string FooterResim
    { get { BaglantiTablosu(); return AyarlarTabloRS["FooterResim"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string DefaultAltBanner
    { get { BaglantiTablosu(); return AyarlarTabloRS["DefaultAltBanner"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string SiparisEmailAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["SiparisEmailAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static bool TopluveTekliResimDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["TopluveTekliResimDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool GoogleUzunlukDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["GoogleUzunlukDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int GoogleTitleUzunluk
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["GoogleTitleUzunluk"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int GoogleTitleUrlUzunluk
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["GoogleTitleUrlUzunluk"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int GoogleDescriptionUzunluk
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["GoogleDescriptionUzunluk"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int GoogleKeywordsUzunluk
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["GoogleKeywordsUzunluk"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool GoogleKisalikDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["GoogleKisalikDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int GoogleTitleKisalik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["GoogleTitleKisalik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int GoogleTitleUrlKisalik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["GoogleTitleUrlKisalik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int GoogleDescriptionKisalik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["GoogleDescriptionKisalik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int GoogleKeywordsKisalik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["GoogleKeywordsKisalik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int WatermakTur
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["WatermakTur"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int WatermakResimGenislik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["WatermakResimGenislik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int WatermakResimYukseklik
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["WatermakResimYukseklik"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int WatermakYukariBosluk
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["WatermakYukariBosluk"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int WatermakSolBosluk
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["WatermakSolBosluk"].ToString()); } }
    //-----------------------------------------------------------------------
    public static string WatermakResimTur
    { get { BaglantiTablosu(); return AyarlarTabloRS["WatermakResimTur"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string MailKontrolEmailAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["MailKontrolEmailAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string IletisimFormuEmailAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["IletisimFormuEmailAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static bool KucukResimWatermakDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["KucukResimWatermakDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int WatermakResimGenislikKucuk
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["WatermakResimGenislikKucuk"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int WatermakResimYukseklikKucuk
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["WatermakResimYukseklikKucuk"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int WatermakYukariBoslukKucuk
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["WatermakYukariBoslukKucuk"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int WatermakSolBoslukKucuk
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["WatermakSolBoslukKucuk"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool GlobalAsaxHtmlSizDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["GlobalAsaxHtmlSizDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int ResimUzerineLogoTur
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["ResimUzerineLogoTur"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int ResimUzerineLogoYukariBosluk
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["ResimUzerineLogoYukariBosluk"].ToString()); } }
    //-----------------------------------------------------------------------
    public static int ResimUzerineLogoSolBosluk
    { get { BaglantiTablosu(); return Convert.ToInt32(AyarlarTabloRS["ResimUzerineLogoSolBosluk"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool AcilisReklamIaOlsunDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["AcilisReklamIaOlsunDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool ResimHrefTitleAltTagDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["ResimHrefTitleAltTagDurum"].ToString()); } }
    //-----------------------------------------------------------------------
    public static bool AnlikCropDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["AnlikCropDurum"].ToString()); } }
    //----------------------------------------------------------------------
    public static bool BadRequestKontrol
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["BadRequestKontrol"].ToString()); } }
    //----------------------------------------------------------------------
    public static bool ResimCiftNoktaDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["ResimCiftNoktaDurum"].ToString()); } }
    //----------------------------------------------------------------------
    public static bool ListeTopluSilmeDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["ListeTopluSilmeDurum"].ToString()); } }
    //----------------------------------------------------------------------
    public static bool SiteDilDurum
    { get { BaglantiTablosu(); return Convert.ToBoolean(AyarlarTabloRS["SiteDilDurum"].ToString()); } }
    //----------------------------------------------------------------------
    public static string SiteDilKoduText
    { get { BaglantiTablosu(); return AyarlarTabloRS["SiteDilKoduText"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string SiteDilKoduSec
    { get { BaglantiTablosu(); return AyarlarTabloRS["SiteDilKoduSec"].ToString(); } }
    //-----------------------------------------------------------------------

    public static string SiteEmailAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["SiteEmailAdresi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string SiteNot
    { get { BaglantiTablosu(); return AyarlarTabloRS["SiteNot"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string MobilSiteismi
    { get { BaglantiTablosu(); return AyarlarTabloRS["MobilSiteismi"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string MobilDescription
    { get { BaglantiTablosu(); return AyarlarTabloRS["MobilDescription"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string MobilKeywords
    { get { BaglantiTablosu(); return AyarlarTabloRS["MobilKeywords"].ToString(); } }
    //-----------------------------------------------------------------------

    public static string MobilGoogleAnalitik
    { get { BaglantiTablosu(); return AyarlarTabloRS["MobilGoogleAnalitik"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string MobilGoogleDogrulamaKodu
    { get { BaglantiTablosu(); return AyarlarTabloRS["MobilGoogleDogrulamaKodu"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string MobilTags
    { get { BaglantiTablosu(); return AyarlarTabloRS["MobilTags"].ToString(); } }
    //-----------------------------------------------------------------------
    public static string WhatsAppLinkAdresi
    { get { BaglantiTablosu(); return AyarlarTabloRS["WhatsAppLinkAdresi"].ToString(); } }
    //-----------------------------------------------------------------------

    //-----------------------------------------------------------------------
    //-----------------------------------------------------------------------
    public static bool MailGonderme(string alici, string konu, string icerik)
    {
        try
        {
            MailMessage msg = new MailMessage();
            bool geridonenveri = true;
            BaglantiTablosu();
            if (AyarlarTabloRS != null)
            {
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpserver", AyarlarTabloRS["epostaHost"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpserverport", AyarlarTabloRS["epostaport"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendusing", 2);
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpauthenticate", 1);
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendusername", AyarlarTabloRS["epostaKullaniciAdi"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendpassword", AyarlarTabloRS["epostaParola"].ToString());
                msg.To = alici.ToString();
                msg.From = AyarlarTabloRS["epostaGonderenAd"].ToString() + "<" + AyarlarTabloRS["epostaKullaniciAdi"].ToString() + ">";
                msg.Subject = konu;
                //---------------------------------------------------------------------------------------------------------------
                string shtml = "";
                shtml += MailStyle("body", "font-size", "12px");
                shtml += MailStyle("table", "font-size", "12px");
                shtml += MailStyle("a", "color", "#1567ba");
                shtml += MailStyle("a:hover", "color", "#87c2ff");
                //---------------------------------------------------------------------------------------------------------------
                msg.Body = (shtml + icerik + "<p>" + AyarlarTabloRS["epostaimza"].ToString()) + "</p>";
                msg.BodyFormat = MailFormat.Html;
                msg.BodyEncoding = Encoding.GetEncoding(1254);
                SmtpMail.SmtpServer = AyarlarTabloRS["epostaHost"].ToString();
                SmtpMail.Send(msg);
            }
            return geridonenveri;
        }
        catch { return false; }
    }
    //----------------------------------------------------------------------------------------

    public static bool KullaniciMailGonderme(string alici, string konu, string icerik, string kullanicimail, string kullaniciadisoyadi)
    {
        try
        {
            MailMessage msg = new MailMessage();
            bool geridonenveri = true;
            BaglantiTablosu();
            if (AyarlarTabloRS != null)
            {
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpserver", AyarlarTabloRS["epostaHost"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpserverport", AyarlarTabloRS["epostaport"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendusing", 2);
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpauthenticate", 1);
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendusername", AyarlarTabloRS["epostaKullaniciAdi"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendpassword", AyarlarTabloRS["epostaParola"].ToString());
                msg.To = alici.ToString();
                msg.From = kullanicimail + "<" + kullaniciadisoyadi + ">";
                msg.Subject = konu;
                //---------------------------------------------------------------------------------------------------------------
                string shtml = "";
                shtml += MailStyle("body", "font-size", "12px");
                shtml += MailStyle("table", "font-size", "12px");
                shtml += MailStyle("a", "color", "#1567ba");
                shtml += MailStyle("a:hover", "color", "#87c2ff");
                //---------------------------------------------------------------------------------------------------------------
                msg.Body = (shtml + icerik + "<p>" + AyarlarTabloRS["epostaimza"].ToString()) + "</p>";
                msg.BodyFormat = MailFormat.Html;
                msg.BodyEncoding = Encoding.GetEncoding(1254);
                SmtpMail.SmtpServer = AyarlarTabloRS["epostaHost"].ToString();
                SmtpMail.Send(msg);
            }
            return geridonenveri;
        }
        catch { return false; }

    }
    //----------------------------------------------------------------------------------------

    public static bool GuvenliMailGonderme(string alici, string konu, string icerik)
    {
        try
        {
            MailMessage msg = new MailMessage();
            bool geridonenveri = true;
            BaglantiTablosu();
            if (AyarlarTabloRS != null)
            {
                string shtml = "";
                shtml += MailStyle("body", "font-size", "12px");
                shtml += MailStyle("table", "font-size", "12px");
                shtml += MailStyle("a", "color", "#1567ba");
                shtml += MailStyle("a:hover", "color", "#87c2ff");
                ////---------------------------------------------------------------------------------------------------------------
                System.Net.Mail.MailAddress gonderen = new System.Net.Mail.MailAddress(AyarlarTabloRS["epostaKullaniciAdi"].ToString(), konu);
                System.Net.Mail.MailAddress alan = new System.Net.Mail.MailAddress(alici.ToString(), konu);
                System.Net.Mail.MailMessage eposta = new System.Net.Mail.MailMessage(gonderen, alan);
                eposta.IsBodyHtml = true;
                eposta.Subject = konu;
                eposta.Body = (shtml + icerik + "<p>" + AyarlarTabloRS["epostaimza"].ToString()) + "</p>";
                eposta.BodyEncoding = Encoding.GetEncoding(1254);
                System.Net.NetworkCredential auth = new System.Net.NetworkCredential(AyarlarTabloRS["epostaKullaniciAdi"].ToString(), AyarlarTabloRS["epostaParola"].ToString());
                System.Net.Mail.SmtpClient SMTP = new System.Net.Mail.SmtpClient();
                SMTP.Host = AyarlarTabloRS["epostaHost"].ToString();
                SMTP.UseDefaultCredentials = false;
                SMTP.Credentials = auth;
                SMTP.DeliveryMethod = System.Net.Mail.SmtpDeliveryMethod.Network;
                SMTP.Send(eposta);

            }
            return geridonenveri;
        }
        catch { return false; }
    }
    //----------------------------------------------------------------------------------------

    public static bool DosyaMailGonderme(FileUpload flp, string alici, string konu, string icerik)
    {
        try
        {
            MailMessage msg = new MailMessage();
            bool geridonenveri = true;
            BaglantiTablosu();
            if (AyarlarTabloRS != null)
            {
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpserver", AyarlarTabloRS["epostaHost"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpserverport", AyarlarTabloRS["epostaport"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendusing", 2);
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpauthenticate", 1);
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendusername", AyarlarTabloRS["epostaKullaniciAdi"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendpassword", AyarlarTabloRS["epostaParola"].ToString());
                msg.To = alici.ToString();
                msg.From = AyarlarTabloRS["epostaGonderenAd"].ToString() + "<" + AyarlarTabloRS["epostaKullaniciAdi"].ToString() + ">";
                msg.Subject = konu;
                //---------------------------------------------------------------------------------------------------------------
                string shtml = "";
                shtml += MailStyle("body", "font-size", "12px");
                shtml += MailStyle("table", "font-size", "12px");
                shtml += MailStyle("a", "color", "#1567ba");
                shtml += MailStyle("a:hover", "color", "#87c2ff");
                //---------------------------------------------------------------------------------------------------------------
                msg.Body = (shtml + icerik + "<p>" + AyarlarTabloRS["epostaimza"].ToString()) + "</p>";
                //---------------------------------------------------------------------------------------------------------------
                try
                {
                    if (flp.HasFile)
                    {
                        string dosya = Fonksiyon.ResimDirekYukleme(flp.PostedFile, Fonksiyon.DosyaKlasoru, "");
                        string tamyol = HttpContext.Current.Server.MapPath(Fonksiyon.DosyaKlasoru + dosya.ToString()).ToString();
                        MailAttachment attach = new MailAttachment(tamyol);
                        msg.Attachments.Add(attach);
                    }

                }
                catch { }
                msg.BodyFormat = MailFormat.Html;
                msg.BodyEncoding = Encoding.GetEncoding(1254);
                SmtpMail.SmtpServer = AyarlarTabloRS["epostaHost"].ToString();
                SmtpMail.Send(msg);
            }
            return geridonenveri;
        }
        catch { return false; }
    }
    //---------------------------------------------------------------------------------------------------------------

    public static bool DosyaMailGonderme(string dosya, string alici, string konu, string icerik)
    {
        try
        {
            MailMessage msg = new MailMessage();
            bool geridonenveri = true;
            BaglantiTablosu();
            if (AyarlarTabloRS != null)
            {
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpserver", AyarlarTabloRS["epostaHost"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpserverport", AyarlarTabloRS["epostaport"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendusing", 2);
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpauthenticate", 1);
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendusername", AyarlarTabloRS["epostaKullaniciAdi"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendpassword", AyarlarTabloRS["epostaParola"].ToString());
                msg.To = alici.ToString();
                msg.From = AyarlarTabloRS["epostaGonderenAd"].ToString() + "<" + AyarlarTabloRS["epostaKullaniciAdi"].ToString() + ">";
                msg.Subject = konu;
                //---------------------------------------------------------------------------------------------------------------
                string shtml = "";
                shtml += MailStyle("body", "font-size", "12px");
                shtml += MailStyle("table", "font-size", "12px");
                shtml += MailStyle("a", "color", "#1567ba");
                shtml += MailStyle("a:hover", "color", "#87c2ff");
                //---------------------------------------------------------------------------------------------------------------
                msg.Body = (shtml + icerik + "<p>" + AyarlarTabloRS["epostaimza"].ToString()) + "</p>";
                //---------------------------------------------------------------------------------------------------------------
                try
                {
                    if (!String.IsNullOrEmpty(dosya))
                    {
                        if (!String.IsNullOrEmpty(dosya))
                        {
                            string tamyol = HttpContext.Current.Server.MapPath(Fonksiyon.DosyaKlasoru + dosya.ToString()).ToString();
                            MailAttachment attach = new MailAttachment(tamyol);
                            msg.Attachments.Add(attach);
                        }
                    }

                }
                catch { }
                msg.BodyFormat = MailFormat.Html;
                msg.BodyEncoding = Encoding.GetEncoding(1254);
                SmtpMail.SmtpServer = AyarlarTabloRS["epostaHost"].ToString();
                SmtpMail.Send(msg);
            }
            return geridonenveri;
        }
        catch { return false; }
    }
    //---------------------------------------------------------------------------------------------------------------

    public static bool KullaniciDosyaMailGonderme(FileUpload flp, string alici, string konu, string icerik, string kullanicimail, string kullaniciadisoyadi)
    {
        try
        {
            MailMessage msg = new MailMessage();
            bool geridonenveri = true;
            BaglantiTablosu();
            if (AyarlarTabloRS != null)
            {
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpserver", AyarlarTabloRS["epostaHost"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpserverport", AyarlarTabloRS["epostaport"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendusing", 2);
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpauthenticate", 1);
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendusername", AyarlarTabloRS["epostaKullaniciAdi"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendpassword", AyarlarTabloRS["epostaParola"].ToString());
                msg.To = alici.ToString();
                msg.From = kullanicimail + "<" + kullaniciadisoyadi + ">";
                msg.Subject = konu;
                //---------------------------------------------------------------------------------------------------------------
                string shtml = "";
                shtml += MailStyle("body", "font-size", "12px");
                shtml += MailStyle("table", "font-size", "12px");
                shtml += MailStyle("a", "color", "#1567ba");
                shtml += MailStyle("a:hover", "color", "#87c2ff");
                //---------------------------------------------------------------------------------------------------------------
                msg.Body = (shtml + icerik + "<p>" + AyarlarTabloRS["epostaimza"].ToString()) + "</p>";
                //---------------------------------------------------------------------------------------------------------------
                try
                {
                    if (flp.HasFile)
                    {
                        string dosya = Fonksiyon.ResimDirekYukleme(flp.PostedFile, Fonksiyon.DosyaKlasoru, "");
                        string tamyol = HttpContext.Current.Server.MapPath(Fonksiyon.DosyaKlasoru + dosya.ToString()).ToString();
                        MailAttachment attach = new MailAttachment(tamyol);
                        msg.Attachments.Add(attach);
                    }

                }
                catch { }
                //---------------------------------------------------------------------------------------------------------------
                msg.BodyFormat = MailFormat.Html;
                msg.BodyEncoding = Encoding.GetEncoding(1254);
                SmtpMail.SmtpServer = AyarlarTabloRS["epostaHost"].ToString();
                SmtpMail.Send(msg);
            }
            return geridonenveri;
        }
        catch { return false; }

    }
    //----------------------------------------------------------------------------------------

    public static bool KullaniciDosyaMailGonderme(string dosya, string alici, string konu, string icerik, string kullanicimail, string kullaniciadisoyadi)
    {
        try
        {
            MailMessage msg = new MailMessage();
            bool geridonenveri = true;
            BaglantiTablosu();
            if (AyarlarTabloRS != null)
            {
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpserver", AyarlarTabloRS["epostaHost"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpserverport", AyarlarTabloRS["epostaport"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendusing", 2);
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/smtpauthenticate", 1);
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendusername", AyarlarTabloRS["epostaKullaniciAdi"].ToString());
                msg.Fields.Add("http://schemas.microsoft.com/cdo/configuration/sendpassword", AyarlarTabloRS["epostaParola"].ToString());
                msg.To = alici.ToString();
                msg.From = kullanicimail + "<" + kullaniciadisoyadi + ">";
                msg.Subject = konu;
                //---------------------------------------------------------------------------------------------------------------
                string shtml = "";
                shtml += MailStyle("body", "font-size", "12px");
                shtml += MailStyle("table", "font-size", "12px");
                shtml += MailStyle("a", "color", "#1567ba");
                shtml += MailStyle("a:hover", "color", "#87c2ff");
                //---------------------------------------------------------------------------------------------------------------
                msg.Body = (shtml + icerik + "<p>" + AyarlarTabloRS["epostaimza"].ToString()) + "</p>";
                //---------------------------------------------------------------------------------------------------------------
                try
                {
                    if (!String.IsNullOrEmpty(dosya))
                    {
                        string tamyol = HttpContext.Current.Server.MapPath(Fonksiyon.DosyaKlasoru + dosya.ToString()).ToString();
                        MailAttachment attach = new MailAttachment(tamyol);
                        msg.Attachments.Add(attach);
                    }

                }
                catch { }
                //---------------------------------------------------------------------------------------------------------------
                msg.BodyFormat = MailFormat.Html;
                msg.BodyEncoding = Encoding.GetEncoding(1254);
                SmtpMail.SmtpServer = AyarlarTabloRS["epostaHost"].ToString();
                SmtpMail.Send(msg);
            }
            return geridonenveri;
        }
        catch { return false; }

    }
    //----------------------------------------------------------------------------------------

    public static string MailStyle(string className, string styleName, string styleProp)
    {
        string geridonenveri = "<style type=\"text/css\">";
        geridonenveri += className + "{" + styleName + ": " + styleProp + ";\n}";
        return geridonenveri + "</style>";
    }
    //----------------------------------------------------------------------------------------

}