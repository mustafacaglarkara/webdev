using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Web;
using System.Web.UI;
using System.Web.UI.WebControls;
public class PanelAyarlar
{

    public static bool GirisKontrol()
    {
        bool kontrol = false;
        if (HttpContext.Current.Session["YFZUserIDV"] == null)
        { HttpContext.Current.Response.Redirect("Default.aspx"); kontrol = true; }
        return kontrol;
    }
    //----------------------------------------------------------------------------------------------

    public static string UploadMailGonderme(string gelenveri, string uzanti)
    {
        gelenveri = "";
        gelenveri += "Merhaba<br><br>";
        gelenveri += "http://" + HttpContext.Current.Request.ServerVariables["HTTP_HOST"] + " sitesine http://" + HttpContext.Current.Request.ServerVariables["HTTP_HOST"] + "" + HttpContext.Current.Request.ServerVariables["URL"] + "?" + HttpContext.Current.Request.ServerVariables["QUERY_STRING"] + " adresine " + HttpContext.Current.Request.ServerVariables["remote_addr"] + " ip numarasından  " + uzanti + " isimli zararlı dosya atılmak istendi.";
        return gelenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string LogSqlInjectionMailGonderme(string gelenveri)
    {
        gelenveri = "";
        gelenveri += "Merhaba<br><br>";
        gelenveri += "http://" + HttpContext.Current.Request.ServerVariables["HTTP_HOST"] + " sitesine http://" + HttpContext.Current.Request.ServerVariables["HTTP_HOST"] + "" + HttpContext.Current.Request.ServerVariables["URL"] + "?" + HttpContext.Current.Request.ServerVariables["QUERY_STRING"] + " adresine " + HttpContext.Current.Request.ServerVariables["remote_addr"] + " ip numarasından  Sql Injection saldırısı yapılmak istendi.";
        return gelenveri;
    }
    //-----------------------------------------------------------------------

    public static string PanelAramaReplace(string gelenveri)
    {
        gelenveri = gelenveri.Replace("Ara=", "");
        gelenveri = gelenveri.Replace("MenuKategori=", "");
        gelenveri = gelenveri.Replace("79CBA1185463850DEDBA31F172F1DC5B", "");
        gelenveri = gelenveri.Replace("6848AE6F8E786062F1B23476C9ECD258", "");
        gelenveri = gelenveri.Replace("AB57FD0432E25D5B3013133A1C910D56", "");
        gelenveri = gelenveri.Replace("FD7030716334C4DB5A83BBBC0F7C7F6F", "");
        gelenveri = gelenveri.Replace("A06B33D1EA28E90733617EC889D4E76E", "");
        gelenveri = gelenveri.Replace("EC5704F0D56945D1E5B8F9A2384A2B4B", "");
        gelenveri = gelenveri.Replace("9508A19D7801E07EE7B7628C31BFDD47", "");
        return gelenveri;
    }
    //-----------------------------------------------------------------------

    public static string PanelAltUrlReplace(string gelenveri)
    {
        gelenveri = gelenveri.Replace("id=", "");
        gelenveri = gelenveri.Replace("79CBA1185463850DEDBA31F172F1DC5B", "");
        gelenveri = gelenveri.Replace("6848AE6F8E786062F1B23476C9ECD258", "");
        gelenveri = gelenveri.Replace("AB57FD0432E25D5B3013133A1C910D56", "");
        gelenveri = gelenveri.Replace("FD7030716334C4DB5A83BBBC0F7C7F6F", "");
        gelenveri = gelenveri.Replace("A06B33D1EA28E90733617EC889D4E76E", "");
        gelenveri = gelenveri.Replace("EC5704F0D56945D1E5B8F9A2384A2B4B", "");
        gelenveri = gelenveri.Replace("9508A19D7801E07EE7B7628C31BFDD47", "");
        return gelenveri;
    }
    //-----------------------------------------------------------------------

    public static string HataMesajReplace(string gelenveri)
    {
        gelenveri = gelenveri.Replace("msj=", "");
        gelenveri = gelenveri.Replace("dhx=", "");
        gelenveri = gelenveri.Replace("79CBA1185463850DEDBA31F172F1DC5B", "");
        gelenveri = gelenveri.Replace("6848AE6F8E786062F1B23476C9ECD258", "");
        gelenveri = gelenveri.Replace("AB57FD0432E25D5B3013133A1C910D56", "");
        gelenveri = gelenveri.Replace("FD7030716334C4DB5A83BBBC0F7C7F6F", "");
        gelenveri = gelenveri.Replace("A06B33D1EA28E90733617EC889D4E76E", "");
        gelenveri = gelenveri.Replace("EC5704F0D56945D1E5B8F9A2384A2B4B", "");
        gelenveri = gelenveri.Replace("9508A19D7801E07EE7B7628C31BFDD47", "");
        return gelenveri;
    }
    //-----------------------------------------------------------------------

    public static string HataMesajCropReplace(string gelenveri)
    {
        gelenveri = gelenveri.Replace("msj=", "");
        gelenveri = gelenveri.Replace("79CBA1185463850DEDBA31F172F1DC5B", "");
        gelenveri = gelenveri.Replace("6848AE6F8E786062F1B23476C9ECD258", "");
        gelenveri = gelenveri.Replace("AB57FD0432E25D5B3013133A1C910D56", "");
        gelenveri = gelenveri.Replace("FD7030716334C4DB5A83BBBC0F7C7F6F", "");
        gelenveri = gelenveri.Replace("A06B33D1EA28E90733617EC889D4E76E", "");
        gelenveri = gelenveri.Replace("EC5704F0D56945D1E5B8F9A2384A2B4B", "");
        gelenveri = gelenveri.Replace("9508A19D7801E07EE7B7628C31BFDD47", "");
        return gelenveri;
    }
    //-----------------------------------------------------------------------

    public static string HataMesajAltReplace(string gelenveri)
    {
        gelenveri = gelenveri.Replace("msj=", "");
        gelenveri = gelenveri.Replace("AltSayfa=", "");
        gelenveri = gelenveri.Replace("AltAra=", "");
        gelenveri = gelenveri.Replace("straltkelime=", "");
        gelenveri = gelenveri.Replace("straltanasayfadurum=", "");
        gelenveri = gelenveri.Replace("stralttur=", "");
        gelenveri = gelenveri.Replace("id=", "");
        gelenveri = gelenveri.Replace("&False", "");
        gelenveri = gelenveri.Replace("&True=", "");
        gelenveri = gelenveri.Replace("Altadd&", "");
        gelenveri = gelenveri.Replace("&&", "&");
        gelenveri = gelenveri.Replace("?1&", "?");
        gelenveri = gelenveri.Replace("tyz=", "");
        gelenveri = gelenveri.Replace("79CBA1185463850DEDBA31F172F1DC5B", "");
        gelenveri = gelenveri.Replace("6848AE6F8E786062F1B23476C9ECD258", "");
        gelenveri = gelenveri.Replace("AB57FD0432E25D5B3013133A1C910D56", "");
        gelenveri = gelenveri.Replace("FD7030716334C4DB5A83BBBC0F7C7F6F", "");
        gelenveri = gelenveri.Replace("A06B33D1EA28E90733617EC889D4E76E", "");
        gelenveri = gelenveri.Replace("EC5704F0D56945D1E5B8F9A2384A2B4B", "");
        gelenveri = gelenveri.Replace("9508A19D7801E07EE7B7628C31BFDD47", "");
        return gelenveri;
    }
    //-----------------------------------------------------------------------

    public static string HataMesajAltGaleriReplace(string gelenveri)
    {
        gelenveri = gelenveri.Replace("tyz=k_", "");
        gelenveri = gelenveri.Replace("tyz=b_", "");
        gelenveri = gelenveri.Replace("AltTitle=", "");
        return gelenveri;
    }
    //-----------------------------------------------------------------------

    public static bool Onay(object value)
    {
        bool geriDon; int deger = Convert.ToInt32(value);
        if (deger == 1) { geriDon = true; }
        else { geriDon = false; }
        return geriDon;
    }
    //-----------------------------------------------------------------------

    public static string VeriRaporlari(String Baslik, int Durum)
    {
        string gelenveri = "";
        if (Ayarlar.VeriRaporDurum == true)
        {
            string SiteAdi = "http://" + HttpContext.Current.Request.ServerVariables["HTTP_HOST"];
            string Yol = HttpContext.Current.Request.ServerVariables["URL"];
            string Devami = HttpContext.Current.Request.ServerVariables["QUERY_STRING"];
            string OncekiADRES = (SiteAdi) + (Yol) + ("?") + (Devami);
            //---------------------------------------------------------------------------------
            String sql = "insert into Tbl_LogKayitlar";
            sql += " (UserID,Email,Bolum,Baslik,Link,Durum,Tarih,KayitTarihi,LogIP) VALUES ";
            sql += "('" + Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]) + "','" + Fonksiyon.SqlTemizle(HttpContext.Current.Session["PKullaniciAdi"].ToString()) + "','" + Fonksiyon.SqlTemizle(HttpContext.Current.Request["Title"]) + "','" + Fonksiyon.SqlTemizle(Baslik) + "','" + Fonksiyon.SqlTemizle(OncekiADRES) + "','" + PanelRaporDurumBelirt(Convert.ToString(Durum)) + "','" + Fonksiyon.TarihFormatiniTersCevir(DateTime.Now.Date.ToString()) + "','" + Fonksiyon.SqlTemizle(DateTime.Now.ToString()) + "','" + Fonksiyon.SqlTemizle(HttpContext.Current.Request.ServerVariables["remote_addr"]) + "')";
            Baglanti.ExecuteQuery(sql);
            //---------------------------------------------------------------------------------
        }
        return gelenveri;
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string LogGenelRaporlari(int UserID, String Email, String Aciklama, int Tur)
    {
        string gelenveri = "";
        if (Ayarlar.LogRaporDurum == true)
        {
            string SiteAdi = "http://" + HttpContext.Current.Request.ServerVariables["HTTP_HOST"];
            string Yol = HttpContext.Current.Request.ServerVariables["URL"];
            string Devami = HttpContext.Current.Request.ServerVariables["QUERY_STRING"];
            string OncekiADRES = (SiteAdi) + (Yol) + ("?") + (Devami);
            //-----------------------------------------------//
            String logAdd = "insert into Tbl_LogGenel (Tarih,UzunTarih,UserID,LogIP,Email,Url,Aciklama,Tur) VALUES ";
            logAdd += "('" + Fonksiyon.TarihFormatiniTersCevir(DateTime.Now.Date.ToString()) + "','" + Fonksiyon.SqlTemizle(DateTime.Now.ToString()) + "'," + Convert.ToInt32(UserID) + ",'" + Fonksiyon.SqlTemizle(HttpContext.Current.Request.ServerVariables["remote_addr"]) + "','" + Fonksiyon.SqlTemizle(Email) + "','" + Fonksiyon.SqlTemizle(OncekiADRES) + "','" + Fonksiyon.SqlTemizle(Aciklama) + "'," + Convert.ToInt32(Tur) + ")";
            Baglanti.ExecuteQuery(logAdd);
            //-----------------------------------------------//
        }
        return gelenveri;
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string PanelRaporDurumBelirt(string Durum)
    {
        switch ((Durum))
        {
            case "1": Durum = "Kaydet"; break;
            case "2": Durum = "Güncelle"; break;
            case "3": Durum = "Silme"; break;
            case "4": Durum = "Sıralama"; break;
            case "5": Durum = "Resim Silme"; break;
            case "6": Durum = "Resim Crop"; break;
            case "7": Durum = "Dosya Silme"; break;
            case "8": Durum = "Liste Silme"; break;
            case "9": Durum = "Liste Güncelle"; break;
            case "10": Durum = "Geri Yükleme"; break;
        }
        return Durum;
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string LogTurBelirt(string Durum)
    {
        switch ((Durum))
        {
            case "1": Durum = "Panel Girişi"; break;
            case "2": Durum = "Panel Hatalı Şifre Girişi"; break;
            case "3": Durum = "Panel Şifremi Unuttum"; break;
            case "4": Durum = "Panel Sql Injection"; break;
            case "5": Durum = "Panel Upload Injection"; break;
            case "6": Durum = "Site Sql Injection"; break;
            case "7": Durum = "Site Upload Injection"; break;
        }
        return Durum;
    }
    //------------------------------------------------------------------------------------------------------------------

    public static void IslemSonuMesaj(string msj, string url)
    {
        string yonlendirme = null;
        string mesaj = Fonksiyon.md5sifreleme(msj);
        if (String.IsNullOrEmpty(url))
            yonlendirme = Agac.PanelAnasayfa[2];
        if (url.Contains("msj="))
            yonlendirme = url.Replace(url.Substring(url.IndexOf("msj="), 36), "msj=" + mesaj);
        else
            if (!url.Contains("?"))
                yonlendirme = url + "?msj=" + mesaj;
            else
                yonlendirme = url + "&msj=" + mesaj;
        HttpContext.Current.Response.Redirect(yonlendirme, false);
    }
    //------------------------------------------------------------------------------------------------------------------

    public static void BilgiVer(string bilgi, Literal MessageLiteral)
    {
        MessageLiteral.Text = "";
        MessageLiteral.Text += "<div class=\"status info\">";
        MessageLiteral.Text += "<p class=\"closestatus\"><a href=\"?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "\" title=\"Kapat\">x</a></p>";
        MessageLiteral.Text += "<p><img src=\"images/default/icons/icon_info.png\" alt=\"Information\" /><span>Bilgi! </span>" + bilgi + ".</p>";
        MessageLiteral.Text += "</div>";
    }
    //------------------------------------------------------------------------------------------------------------------

    public static void BilgiEkle(string bilgi, Literal MessageLiteral)
    {
        MessageLiteral.Text += "<div class=\"status info\">";
        MessageLiteral.Text += "<p class=\"closestatus\"><a href=\"?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "\" title=\"Kapat\">x</a></p>";
        MessageLiteral.Text += "<p><img src=\"images/default/icons/icon_info.png\" alt=\"Information\" /><span>Bilgi! </span>" + bilgi + ".</p>";
        MessageLiteral.Text += "</div>";
    }
    //------------------------------------------------------------------------------------------------------------------

    public static void MailUyari(string bilgi, Literal MessageLiteral)
    {
        MessageLiteral.Text = "";
        MessageLiteral.Text += "<div class=\"status error\">";
        MessageLiteral.Text += "<p class=\"closestatus\"><a href=\"?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "\" title=\"Kapat\">x</a></p>";
        MessageLiteral.Text += "<p><img src=\"images/default/icons/icon_info.png\" alt=\"Information\" /><span>Bilgi! </span>" + bilgi + ".</p>";
        MessageLiteral.Text += "</div>";
    }
    //------------------------------------------------------------------------------------------------------------------

    public static void GetMsj(string msj, Literal lbl)
    {
        string MsjMD5 = msj;
        string bilgiMsj = null;
        string Status = "";
        //------------------------------------------------------------------------------------------------------------------
        if (MsjMD5 == Fonksiyon.md5sifreleme("GB")) { bilgiMsj = "Tebrikler güncelleme başarılı."; Status = "Basari"; }
        else if (MsjMD5 == Fonksiyon.md5sifreleme("GH")) { bilgiMsj = "Güncelleme yapılırken bir hata oluştu."; Status = "Hata"; }
        else if (MsjMD5 == Fonksiyon.md5sifreleme("KB")) { bilgiMsj = "Tebrikler kayıt başarılı."; Status = "Basari"; }
        else if (MsjMD5 == Fonksiyon.md5sifreleme("KH")) { bilgiMsj = "Kayıt yapılırken bir hata oluştu."; Status = "Hata"; }
        else if (MsjMD5 == Fonksiyon.md5sifreleme("SB")) { bilgiMsj = "Tebrikler silme işlemi başarılı."; Status = "Basari"; }
        else if (MsjMD5 == Fonksiyon.md5sifreleme("SH")) { bilgiMsj = "Silme yapılırken bir hata oluştu."; Status = "Hata"; }
        else if (MsjMD5 == Fonksiyon.md5sifreleme("RH")) { bilgiMsj = "Bu Resmi veya Dosyayı Sistem Kabul Etmiyor."; Status = "Hata"; }

        //------------------------------------------------------------------------------------------------------------------
        Literal ltrl = lbl;
        if (Status == "Basari")
        {
            ltrl.Text += "<div class=\"status success\">";
            ltrl.Text += "<p class=\"closestatus\"><a href=\"?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "\" title=\"Kapat\">x</a></p>";
            ltrl.Text += "<p><img src=\"images/default/icons/icon_success.png\" alt=\"Success\" /><span>Başarı! </span>" + bilgiMsj + ".</p>";
            ltrl.Text += "</div>";
        }
        if (Status == "Bilgi")
        {
            ltrl.Text += "<div class=\"status info\">";
            ltrl.Text += "<p class=\"closestatus\"><a href=\"?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "\" title=\"Kapat\">x</a></p>";
            ltrl.Text += "<p><img src=\"images/default/icons/icon_info.png\" alt=\"Information\" /><span>Bilgi! </span>" + bilgiMsj + ".</p>";
            ltrl.Text += "</div>";
        }
        if (Status == "Hata")
        {
            ltrl.Text += "<div class=\"status error\">";
            ltrl.Text += "<p class=\"closestatus\"><a href=\"?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "\" title=\"Kapat\">x</a></p>";
            ltrl.Text += "<p><img src=\"images/default/icons/icon_error.png\" alt=\"Error\" /><span>Hata! </span>" + bilgiMsj + ".</p>";
            ltrl.Text += "</div>";
        }
        if (Status == "Uyari")
        {
            ltrl.Text += "<div class=\"status warning\">";
            ltrl.Text += "<p class=\"closestatus\"><a href=\"?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "\" title=\"Kapat\">x</a></p>";
            ltrl.Text += "<p><img src=\"images/default/icons/icon_warning.png\" alt=\"Warning\" /><span>Uyarı! </span>" + bilgiMsj + ".</p>";
            ltrl.Text += "</div>";
        }
    }
    //------------------------------------------------------------------------------------------------------------

    public static string GenelResimGaleriSec(string gelenveri)
    {
        string shtml = "<a href=\"" + Fonksiyon.ResimKlasoru + "b_" + gelenveri + "\" class=\"lightbox\"><img src=\"" + Fonksiyon.ResimKlasoru + "k_" + gelenveri + "\" style=\"padding:2px; background-color:Black\" border=\"0\"></a>";
        return shtml;
    }
    //----------------------------------------------------------------------------------------------

    public static string GenelResimGaleriKucukSec(string gelenveri)
    {
        string shtml = "<a href=\"" + Fonksiyon.ResimKlasoru + "k_" + gelenveri + "\" class=\"lightbox\"><img src=\"" + Fonksiyon.ResimKlasoru + "k_" + gelenveri + "\" style=\"padding:2px; background-color:Black\" border=\"0\"></a>";
        return shtml;
    }
    //----------------------------------------------------------------------------------------------

    public static string GenelKucukResimGaleriSec(string gelenveri)
    {
        string shtml = "<a href=\"" + Fonksiyon.ResimKlasoru + "k_" + gelenveri + "\" class=\"lightbox\"><img src=\"" + Fonksiyon.ResimKlasoru + "k_" + gelenveri + "\" style=\"padding:2px; background-color:Black\" border=\"0\"></a>";
        return shtml;
    }
    //----------------------------------------------------------------------------------------------

    public static string GenelResimBuyukGaleriSec(string gelenveri)
    {
        string shtml = "<a href=\"" + Fonksiyon.ResimKlasoru + "b_" + gelenveri + "\" class=\"lightbox\"><img src=\"" + Fonksiyon.ResimKlasoru + "b_" + gelenveri + "\" style=\"padding:2px; background-color:Black; max-height:100px; max-width:100px; \" border=\"0\"></a>";
        return shtml;
    }
    //----------------------------------------------------------------------------------------------

    public static string GenelResimDirekGaleriSec(string gelenveri)
    {
        string shtml = "<a href=\"" + Fonksiyon.ResimKlasoru + gelenveri + "\" class=\"lightbox\"><img src=\"" + Fonksiyon.ResimKlasoru + gelenveri + "\" style=\"padding:2px; background-color:Black; max-height:100px; max-width:100px; \" border=\"0\"></a>";
        return shtml;
    }
    //----------------------------------------------------------------------------------------------

    public static string GenelDetayBuyukSec(string gelenveri)
    {
        string shtml = "";
        if (gelenveri != "")
        { shtml = "<a href=\"" + Fonksiyon.ResimKlasoru + "b_" + gelenveri + "\" class=\"lightbox\"><img src=\"" + Fonksiyon.ResimKlasoru + "b_" + gelenveri + "\" style=\"padding:2px; background-color:Black;  \" border=\"0\"></a>"; }
        return shtml;
    }
    //----------------------------------------------------------------------------------------------

    public static string ListeGaleriSec(string gelenveri)
    {
        string shtml = "";
        if (gelenveri != "")
        { shtml = "<a href=\"" + Fonksiyon.ResimKlasoru + "b_" + gelenveri + "\" class=\"lightbox\"><img src=\"" + Fonksiyon.ResimKlasoru + "k_" + gelenveri + "\" style=\"padding:2px; background-color:Black; max-height:50px; max-width:50px; \" border=\"0\"></a>"; }
        return shtml;
    }
    //----------------------------------------------------------------------------------------------

    public static string ListeGaleriBuyukSec(string gelenveri)
    {
        string shtml = "";
        if (gelenveri != "")
        { shtml = "<a href=\"" + Fonksiyon.ResimKlasoru + "b_" + gelenveri + "\" class=\"lightbox\"><img src=\"" + Fonksiyon.ResimKlasoru + "b_" + gelenveri + "\" style=\"padding:2px; background-color:Black; max-height:50px; max-width:50px; \" border=\"0\"></a>"; }
        return shtml;
    }
    //----------------------------------------------------------------------------------------------

    public static string ListeDirekGaleriSec(string gelenveri)
    {
        string shtml = "";
        if (gelenveri != "")
        { shtml = "<a href=\"" + Fonksiyon.ResimKlasoru + gelenveri + "\" class=\"lightbox\"><img src=\"" + Fonksiyon.ResimKlasoru + gelenveri + "\" style=\"padding:2px; background-color:Black; max-height:50px; max-width:50px; \" border=\"0\"></a>"; }
        return shtml;
    }
    //----------------------------------------------------------------------------------------------

    public static string GenelFlashSec(string gelenveri)
    {
        string shtml = "";
        shtml += "<object classid=\"clsid:d27cdb6e-ae6d-11cf-96b8-444553540000\" codebase=\"http://download.macromedia.com/pub/shockwave/cabs/flash/swflash.cab#version=7,0,0,0\" max-width=\"180\" max-height=\"150\" id=\"Ad\" align=\"middle\">";
        shtml += "<param name=\"allowScriptAccess\" value=\"sameDOMAIN\" />";
        shtml += "<param name=\"movie\" value=\"" + Fonksiyon.ResimKlasoru + "" + gelenveri + "\" />";
        shtml += "<param name=\"wmode\" value=\"transparent\">";
        shtml += "<param name=\"quality\" value=\"High\" />";
        shtml += "<embed src=\"" + Fonksiyon.ResimKlasoru + "" + gelenveri + "\" quality=\"High\" bgcolor=\"#ffffff\" max-width=\"180\" max-height=\"150\" name=\"Ad\" align=\"middle\" allowscriptaccess=\"sameDOMAIN\" type=\"application/x-shockwave-flash\" pluginspage=\"http://www.macromedia.com/go/getflashplayer\" />";
        shtml += "</object>";
        return shtml;
    }
    //----------------------------------------------------------------------------------------------

    public static string GenelDosyaSec(string gelenveri)
    {
        string shtml = "<a href=\"" + Fonksiyon.DosyaKlasoru + "" + gelenveri + "\" class=\"ttrresimduzenleme\" target=\"_blank\">Dosyayı İndirmek İçin Tıklayınız.</a>";
        return shtml;
    }
    //----------------------------------------------------------------------------------------------

    public static string GenelVideoDosyaSec(string gelenveri)
    {
        string shtml = "Video Dosya : " + gelenveri + "";
        return shtml;
    }
    //----------------------------------------------------------------------------------------------

    public static string GoogleKarakterKontrolDurum(string verititle, string veriurl, string veridescription, string verikeywords)
    {
        string geridonenveri = "";
        //---------------------------------------------------------------------------------
        if (Ayarlar.GoogleUzunlukDurum == true)
        {
            if (verititle.Length > Ayarlar.GoogleTitleUzunluk)
            { geridonenveri = "Başlık Title " + Ayarlar.GoogleTitleUzunluk + " Karakterden Uzun Olamaz. "; }
            //------------------------------------------------------------------------
            else if (veriurl.Length > Ayarlar.GoogleTitleUrlUzunluk)
            { geridonenveri = "Başlık Url " + Ayarlar.GoogleTitleUrlUzunluk + " Karakterden Uzun Olamaz. "; }
            //------------------------------------------------------------------------
            else if (veridescription.Length > Ayarlar.GoogleDescriptionUzunluk)
            { geridonenveri = "Description " + Ayarlar.GoogleDescriptionUzunluk + " Karakterden Uzun Olamaz. "; }
            //------------------------------------------------------------------------
            else if (verikeywords.Length > Ayarlar.GoogleKeywordsUzunluk)
            { geridonenveri = "Keywords " + Ayarlar.GoogleKeywordsUzunluk + " Karakterden Uzun Olamaz. "; }
        }
        //---------------------------------------------------------------------------------

        //---------------------------------------------------------------------------------
        if (Ayarlar.GoogleKisalikDurum == true)
        {
            if (verititle.Length < Ayarlar.GoogleTitleKisalik)
            { geridonenveri = "Başlık Title " + Ayarlar.GoogleTitleKisalik + " Karakterden Kısa Olamaz. "; }
            //------------------------------------------------------------------------
            else if (veriurl.Length < Ayarlar.GoogleTitleUrlKisalik)
            { geridonenveri = "Başlık Url " + Ayarlar.GoogleTitleUrlKisalik + " Karakterden Kısa Olamaz. "; }
            //------------------------------------------------------------------------
            else if (veridescription.Length < Ayarlar.GoogleDescriptionKisalik)
            { geridonenveri = "Description " + Ayarlar.GoogleDescriptionKisalik + " Karakterden Kısa Olamaz. "; }
            //------------------------------------------------------------------------
            else if (verikeywords.Length < Ayarlar.GoogleKeywordsKisalik)
            { geridonenveri = "Keywords " + Ayarlar.GoogleKeywordsKisalik + " Karakterden Kısa Olamaz. "; }
        }
        //---------------------------------------------------------------------------------
        return geridonenveri;
        //---------------------------------------------------------------------------------
    }
    //----------------------------------------------------------------------------------------------

}