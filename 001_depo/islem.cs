using System;
using System.Data;
using System.Collections.Generic;
using System.Linq;
using System.Web;
using System.Xml;
using System.Data.OleDb;
using System.Web.UI;
using System.Web.UI.WebControls;
using System.Net;
public class islem
{
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------
    //--------------------------Kısayol İşlemleri Başlangıç----------------------------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------

    public static string GlobalRewRiteLink;
    //----------------------------------------------------------------------------------------------

    public static string GlobalRequestSiz()
    {
        try
        {
            //--------------------------------------------------------------------------------
            if (Ayarlar.BadRequestKontrol == true)
            { Fonksiyon.BadRequestKontrol(); }
            //--------------------------------------------------------------------------------
            string katlink = Fonksiyon.SqlInjectionTemizle(HttpContext.Current.Request.RawUrl.Replace("/", "").Substring(0, HttpContext.Current.Request.RawUrl.Substring(1).IndexOf("/")));
            //--------------------------------------------------------------------------------
            return katlink;
            //--------------------------------------------------------------------------------
        }
        catch { return null; }
    }
    //-------------------------------------------------------------------------------------------

    public static string GlobalRequestSizSadeceUrl(int durum)
    {
        try
        {
            string katlink = "";
            if (Ayarlar.GlobalAsaxHtmlSizDurum == false)
            { katlink = Fonksiyon.SqlInjectionTemizle(HttpContext.Current.Request.RawUrl.Replace("/", "").Substring(0, HttpContext.Current.Request.RawUrl.Substring(1).IndexOf("/"))); }
            else
            {
                //--------------------------------------------------------------------------------
                int urlsayi = HttpContext.Current.Request.RawUrl.Substring(1).IndexOf("/");
                //--------------------------------------------------------------------------------
                katlink = Fonksiyon.SqlInjectionTemizle(HttpContext.Current.Request.RawUrl.Replace("/", "").Substring(0, urlsayi));
                //--------------------------------------------------------------------------------
                string kontroldeger = null;
                //--------------------------------------------------------------------------------
                if (durum == 0)
                {
                    try
                    { kontroldeger = Fonksiyon.SqlInjectionTemizle(HttpContext.Current.Request.RawUrl.Replace("/", "").Substring(0, urlsayi + 1)); }
                    catch { }
                }
                //--------------------------------------------------------------------------------
                if (!String.IsNullOrEmpty(kontroldeger))
                {
                    if (katlink.Length < kontroldeger.Length)
                    { HttpContext.Current.Response.Redirect("/404.html"); }
                }
                //--------------------------------------------------------------------------------
            }
            return katlink;
            //--------------------------------------------------------------------------------
        }
        catch { return null; }
    }
    //-------------------------------------------------------------------------------------------
    
    public static bool AnlikCropDurum(int tur)
    {
        //----------------------------------------------------------
        if (Ayarlar.GlobalAsaxHtmlSizDurum == true)
        {
            Baglanti.ExecuteQuery("update Tbl_SiteAyarlari set AnlikCropDurum=" + Convert.ToInt32(tur) + "");
            Ayarlar.AyarlarTabloRS = null;
            islem.GlobalRewRiteLink = "GlobalTemizle";
        }
        return true;
    }
    //--------------------------------------------------------------------------------------------
    
    public static string VeriGetirSec(string gelenveri, string tabloadi, string alanadi, string aranacakalan)
    {
        string geridonenveri = null;
        DataRow rs = Baglanti.GetDataRow("select " + Fonksiyon.SqlTemizle(alanadi) + " from Tbl_" + Fonksiyon.SqlTemizle(tabloadi) + " where " + Fonksiyon.SqlTemizle(aranacakalan) + "='" + Fonksiyon.SqlInjectionTemizle(gelenveri) + "'");
        if (rs != null)
        { geridonenveri = rs[0].ToString(); }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static bool ResimSilSec(string gelenveri, string tabloadi, string alanadi, string aranacakalan, string resimyolu)
    {
        bool geridonenveri = true;
        if (Ayarlar.ResimSilmeDurum == true)
        {
            DataRow rs = Baglanti.GetDataRow("select " + Fonksiyon.SqlTemizle(alanadi) + " from Tbl_" + Fonksiyon.SqlTemizle(tabloadi) + " where " + Fonksiyon.SqlTemizle(aranacakalan) + "='" + Fonksiyon.SqlInjectionTemizle(gelenveri) + "'");
            if (rs != null)
            {
                try { System.IO.File.Delete(HttpContext.Current.Server.MapPath(resimyolu + "b_" + rs[0].ToString())); }
                catch { }
                try { System.IO.File.Delete(HttpContext.Current.Server.MapPath(resimyolu + "k_" + rs[0].ToString())); }
                catch { }
                try { System.IO.File.Delete(HttpContext.Current.Server.MapPath(resimyolu + "o_" + rs[0].ToString())); }
                catch { }
                try { System.IO.File.Delete(HttpContext.Current.Server.MapPath(resimyolu + rs[0].ToString())); }
                catch { }
                //---------------------------------------------------------------------------------------------------------------
                try { System.IO.File.Delete(HttpContext.Current.Server.MapPath(Fonksiyon.TempCropKlasoru + "b_" + rs[0].ToString())); }
                catch { }
                try { System.IO.File.Delete(HttpContext.Current.Server.MapPath(Fonksiyon.TempCropKlasoru + "k_" + rs[0].ToString())); }
                catch { }
                try { System.IO.File.Delete(HttpContext.Current.Server.MapPath(Fonksiyon.TempCropKlasoru + "o_" + rs[0].ToString())); }
                catch { }
                try { System.IO.File.Delete(HttpContext.Current.Server.MapPath(Fonksiyon.TempCropKlasoru + rs[0].ToString())); }
                catch { }
                //---------------------------------------------------------------------------------------------------------------
                try { System.IO.File.Delete(HttpContext.Current.Server.MapPath(Fonksiyon.TempKlasoru + "b_" + rs[0].ToString())); }
                catch { }
                try { System.IO.File.Delete(HttpContext.Current.Server.MapPath(Fonksiyon.TempKlasoru + "k_" + rs[0].ToString())); }
                catch { }
                try { System.IO.File.Delete(HttpContext.Current.Server.MapPath(Fonksiyon.TempKlasoru + "o_" + rs[0].ToString())); }
                catch { }
                try { System.IO.File.Delete(HttpContext.Current.Server.MapPath(Fonksiyon.TempKlasoru + rs[0].ToString())); }
                catch { }
            }
        }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string ToplamKayitSec(string gelenveri, string tabloadi, string aranacakalan)
    {
        string geridonenveri = null;
        DataRow rs = Baglanti.GetDataRow("select count(id) as toplamsayi from Tbl_" + Fonksiyon.SqlTemizle(tabloadi) + " where " + Fonksiyon.SqlTemizle(aranacakalan) + "='" + Fonksiyon.SqlInjectionTemizle(gelenveri) + "'");
        if (rs != null)
        { geridonenveri = rs[0].ToString(); }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static int ListelemeSayisiSec(string gelenveri)
    {
        //---------------------------------------------------------------------------------
        int geridonenveri = 0;
        //---------------------------------------------------------------------------------
        DataRow rs = Baglanti.GetDataRow("select top 1 ListemeSayisi,SayfaTurID from Tbl_Genel_Kategori where id=" + Convert.ToInt32(gelenveri) + " and ListemeSayisi>0 ");
        if (rs != null)
        {
            //------------------------------------------------------------------------------
            geridonenveri = Convert.ToInt32(rs["ListemeSayisi"].ToString()); 
            //------------------------------------------------------------------------------
            DataRow rsl = Baglanti.GetDataRow("select top 1 ListeSayfalamaSayisi from Tbl_Zm_SayfaTurleri where id=" + Convert.ToInt32(rs["SayfaTurID"].ToString()) + " and ListeSayfalamaSayisi>0 ");
            if (rsl != null)
            { geridonenveri = Convert.ToInt32(rsl["ListeSayfalamaSayisi"].ToString());  }
            //------------------------------------------------------------------------------
        }
        //---------------------------------------------------------------------------------
        return geridonenveri;
        //---------------------------------------------------------------------------------
    }
    //----------------------------------------------------------------------------------------------

    public static string VeriUrlYeniKayitSec(string TabloAdi, string gelenveri, string aranacakalan)
    {
        DataTable rs = Baglanti.GetDataTable("select top 1 id from Tbl_" + Fonksiyon.SqlInjectionTemizle(TabloAdi) + " where " + Fonksiyon.SqlTemizle(aranacakalan) + "='" + Fonksiyon.SeoReplace(gelenveri) + "'");
        if (rs.Rows.Count > 0)
        {
            if (Ayarlar.SeoUrlKontrolDurum == false)
            {
                if (HttpContext.Current.Session["UrlKontrol"] == null)
                { HttpContext.Current.Session["UrlKontrol"] = "ilk"; gelenveri = VeriUrlYeniKayitSec(TabloAdi, gelenveri + "-" + Fonksiyon.UrlRandomSayi.ToString(), aranacakalan); }
                else
                { gelenveri = VeriUrlYeniKayitSec(TabloAdi, gelenveri + Fonksiyon.UrlRandomSayi.ToString(), aranacakalan); }
            }
            else
            { gelenveri = "VeriUrlYeniKayitSec"; }
        }
        return gelenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string VeriUrlDuzenleSec(string TabloAdi, string gelenveri, string id, string aranacakalan)
    {
        DataTable rs = Baglanti.GetDataTable("select top 1  id from Tbl_" + Fonksiyon.SqlInjectionTemizle(TabloAdi) + " where " + Fonksiyon.SqlTemizle(aranacakalan) + "='" + Fonksiyon.SeoReplace(gelenveri) + "' and id<>" + Convert.ToInt32(id) + "");
        if (rs.Rows.Count > 0)
        {
            if (Ayarlar.SeoUrlKontrolDurum == false)
            {
                if (HttpContext.Current.Session["UrlKontrol"] == null)
                { HttpContext.Current.Session["UrlKontrol"] = "ilk"; gelenveri = VeriUrlDuzenleSec(TabloAdi, gelenveri + "-" + id.ToString(), id, aranacakalan); }
                else
                { gelenveri = VeriUrlDuzenleSec(TabloAdi, gelenveri + id.ToString(), id, aranacakalan); }
            }
            else
            { gelenveri = "VeriUrlDuzenleSec"; }
        }
        return gelenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string UstKategorideYoksaEnUstKategoriID(string gelenveri, string kontrolalan)
    {
        string geridonenveri = gelenveri;
        //----------------------------------------------------------------------------------------
        string AtkifID = CokluEnUstKategoriIDSec(Convert.ToInt32(gelenveri));
        string AtkifUID = VeriGetirSec(gelenveri, "Genel_Kategori", "UstKategoriID", "id");
        //----------------------------------------------------------------------------------------
        if (!String.IsNullOrEmpty(VeriGetirSec(AtkifUID, "Genel_Kategori", kontrolalan, "id")))
        { gelenveri = AtkifUID; }
        else if (!String.IsNullOrEmpty(VeriGetirSec(AtkifID, "Genel_Kategori", kontrolalan, "id")))
        { gelenveri = AtkifID; }
        //----------------------------------------------------------------------------------------
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static bool CokluEnUstMenuDetayTekliDurumSec(int UstKategoriID)
    {
        bool geridonenveri = false;
        DataRow rs = Baglanti.GetDataRow("Select TOP 1 MenuDetayTekliDurum,UstKategoriID From Tbl_Genel_Kategori Where id=" + Convert.ToInt32(UstKategoriID) + "");
        if (rs != null)
        {
            geridonenveri = Convert.ToBoolean(rs["MenuDetayTekliDurum"].ToString());
            if (rs["UstKategoriID"].ToString() != "0")
            { geridonenveri = CokluEnUstMenuDetayTekliDurumSec(Convert.ToInt32(rs["UstKategoriID"].ToString())); }
        }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string CokluEnUstKategoriSec(int UstKategoriID)
    {
        string geridonenveri = "";
        DataRow rs = Baglanti.GetDataRow("Select TOP 1 Kategori,UstKategoriID From Tbl_Genel_Kategori Where id=" + Convert.ToInt32(UstKategoriID) + "");
        if (rs != null)
        {
            geridonenveri = rs["Kategori"].ToString();
            if (rs["UstKategoriID"].ToString() != "0")
            { geridonenveri = CokluEnUstKategoriSec(Convert.ToInt32(rs["UstKategoriID"].ToString())); }
        }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string CokluEnUstKategoriIDSec(int UstKategoriID)
    {
        string geridonenveri = "";
        DataRow rs = Baglanti.GetDataRow("Select TOP 1 id,UstKategoriID FROM Tbl_Genel_Kategori Where id=" + Convert.ToInt32(UstKategoriID) + "");
        if (rs != null)
        {
            geridonenveri = rs["id"].ToString();
            if (rs["UstKategoriID"].ToString() != "0")
            { geridonenveri = CokluEnUstKategoriIDSec(Convert.ToInt32(rs["UstKategoriID"].ToString())); }
        }
        return geridonenveri;
    }
    //--------------------------------------------------------------------------------------------

    public static string CssNoCokluKategoriSec(int UstKategoriID, string cssno)
    {
        string geridonenveri = "";
        DataRow rs = Baglanti.GetDataRow("Select TOP 1 Kategori,UstKategoriID,CssNo From Tbl_Genel_Kategori Where id=" + Convert.ToInt32(UstKategoriID) + "");
        if (rs != null)
        {
            geridonenveri = rs["Kategori"].ToString();
            if (rs["CssNo"].ToString() != cssno.ToString())
            { geridonenveri = CssNoCokluKategoriSec(Convert.ToInt32(rs["UstKategoriID"].ToString()), cssno); }
        }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string CssNoCokluKategoriIDSec(int UstKategoriID, string cssno)
    {
        string geridonenveri = "";
        DataRow rs = Baglanti.GetDataRow("Select TOP 1 id,UstKategoriID,CssNo FROM Tbl_Genel_Kategori Where id=" + Convert.ToInt32(UstKategoriID) + "");
        if (rs != null)
        {
            geridonenveri = rs["id"].ToString();
            if (rs["CssNo"].ToString() != cssno.ToString())
            { geridonenveri = CssNoCokluKategoriIDSec(Convert.ToInt32(rs["UstKategoriID"].ToString()), cssno); }
        }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string PanelKategoriSec(int catid)
    {
        string geridonenveri = "";
        DataTable rs = Baglanti.GetDataTable("SELECT top 2 Tbl_Genel_Kategori.UstKategoriID, Tbl_Genel_Kategori.Kategori, Tbl_Genel_Kategori.CssNo FROM Tbl_Genel_Kategori INNER JOIN Tbl_Genel_Kayitlar ON Tbl_Genel_Kategori.id = Tbl_Genel_Kayitlar.CatID where Tbl_Genel_Kategori.id=" + System.Convert.ToInt32(catid) + "");
        if (rs.Rows.Count > 0)
        {
            if (rs.Rows.Count == 1)
            {
                int ccsnosayi = Convert.ToInt32(rs.Rows[0]["CssNo"]) - 1;
                if (ccsnosayi == 0)
                {
                    DataRow rsu = Baglanti.GetDataRow("select top 1 Kategori from Tbl_Genel_Kategori where CssNo=" + System.Convert.ToInt32(ccsnosayi) + " and id=" + System.Convert.ToInt32(rs.Rows[0]["UstKategoriID"].ToString()) + "");
                    if (rsu != null)
                    { geridonenveri = rsu["Kategori"].ToString(); }
                }
                else
                { geridonenveri = rs.Rows[0]["Kategori"].ToString(); }
            }
            else
            { geridonenveri = rs.Rows[0]["Kategori"].ToString(); }
            //-----------------------------------------------------
            return Fonksiyon.IlkHarflerBuyuk(geridonenveri);
        }
        else
        { return geridonenveri; }
    }
    //----------------------------------------------------------------------------------------------

    public static string CokluKategoriSec(string gelenveri)
    {
        string geridonenveri = "";
        DataRow rs = Baglanti.GetDataRow("select Kategori,UstKategoriID from Tbl_Genel_Kategori where id=" + System.Convert.ToInt32(gelenveri) + "");
        if (rs != null)
        {
            geridonenveri = rs["Kategori"].ToString() + " / ";
            if (rs["UstKategoriID"].ToString() != "0")
            { geridonenveri = CokluKategoriSec(rs["UstKategoriID"].ToString()) + geridonenveri; }
        }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string CokluKategoriIDSec(string gelenveri)
    {
        string donecek = gelenveri;
        DataTable rs = Baglanti.GetDataTable("select id,Kategori from Tbl_Genel_Kategori where UstKategoriID=" + System.Convert.ToInt32(gelenveri) + "");
        for (int i = 0; i <= rs.Rows.Count - 1; i++)
        { donecek += "," + CokluKategoriIDSec(rs.Rows[i]["id"].ToString()); }
        return donecek;
    }
    //----------------------------------------------------------------------------------------------

    public static string SiteCokluKategoriIDSec(string gelenveri)
    {
        string donecek = gelenveri;
        DataTable rs = Baglanti.GetDataTable("select id,UstKategoriID from Tbl_Genel_Kategori where id=" + System.Convert.ToInt32(gelenveri) + "");
        for (int i = 0; i <= rs.Rows.Count - 1; i++)
        { donecek += "," + SiteCokluKategoriIDSec(rs.Rows[i]["UstKategoriID"].ToString()); }
        return donecek;
    }
    //--------------------------------------------------------------------------------------------

    public static string SitemapKategoriKontrolSec(string gelenveri)
    {
        string geridonenveri = "";
        DataRow rs = Baglanti.GetDataRow("select id,UstKategoriID from Tbl_Genel_Kategori where id=" + System.Convert.ToInt32(gelenveri) + " and Onay=1 and SilDurum=0 ");
        if (rs != null)
        {
            geridonenveri = rs["id"].ToString();
            if (rs["UstKategoriID"].ToString() != "0")
            { geridonenveri = SitemapKategoriKontrolSec(rs["UstKategoriID"].ToString()) + geridonenveri; }
        }
        else
        { geridonenveri = ""; }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    //----------------------------------------------------------------------------------------------
    //--------------------------Kısayol İşlemleri Son----------------------------------------------
    //----------------------------------------------------------------------------------------------


    //----------------------------------------------------------------------------------------------
    //--------------------------Agaç İşlemleri Başlangıç-----------------------------------------
    //----------------------------------------------------------------------------------------------

    public static void TekliAgacNerdeyim(string gelenveri, string title, Literal ltrlAgac, Page sayfa)
    {
        Agac.Add(Fonksiyon.IlkHarflerBuyuk(gelenveri), "");
        Agac.RegisterAgacSite(ltrlAgac);
        Fonksiyon.SetTitle(Fonksiyon.IlkHarflerBuyuk(title), sayfa);
    }
    //----------------------------------------------------------------------------------------------

    public static void CokluAgacNerdeyim(string gelenveri, string title, int kategoriid, Literal ltrlAgac, Page sayfa)
    {
        string urllink = "";
        DataTable rsn = Baglanti.GetDataTable("Select id,Link,Kategori,SayfaTurID,KategoriSiteMapGorunmesin from Tbl_Genel_Kategori Where id in (" + islem.SiteCokluKategoriIDSec(kategoriid.ToString()) + ") and Onay=1 order by CssNo asc");
        for (int i = 0; i <= rsn.Rows.Count - 1; i++)
        {
            if (!String.IsNullOrEmpty(rsn.Rows[i]["Link"].ToString()))
            { urllink = rsn.Rows[i]["Link"].ToString(); }
            else if (Convert.ToBoolean(rsn.Rows[i]["KategoriSiteMapGorunmesin"].ToString()) == true)
            { urllink = "javascript:;"; }
            else
            { urllink = (islem.GlobalSiteLinkSec(rsn.Rows[i]["SayfaTurID"].ToString(), rsn.Rows[i]["id"].ToString())); }
            //----------------------------------------------------------------------------------------------
            Agac.Add(Fonksiyon.IlkHarflerBuyuk(rsn.Rows[i]["Kategori"].ToString()), urllink);
        }
        Agac.RegisterAgacSite(ltrlAgac);
        //----------------------------------------------------------------------------------
        Fonksiyon.SetTitle(Fonksiyon.IlkHarflerBuyuk(title), sayfa);
        //----------------------------------------------------------------------------------
    }
    //----------------------------------------------------------------------------------------------

    public static void CokluAgacDetayNerdeyim(string gelenveri, string title, int kategoriid, Literal ltrlAgac, Page sayfa)
    {
        string urllink = "";
        DataTable rsn = Baglanti.GetDataTable("Select id,Link,Kategori,SayfaTurID,KategoriSiteMapGorunmesin from Tbl_Genel_Kategori Where id in (" + islem.SiteCokluKategoriIDSec(kategoriid.ToString()) + ") and Onay=1 order by CssNo asc");
        for (int i = 0; i <= rsn.Rows.Count - 1; i++)
        {
            if (!String.IsNullOrEmpty(rsn.Rows[i]["Link"].ToString()))
            { urllink = rsn.Rows[i]["Link"].ToString(); }
            else if (Convert.ToBoolean(rsn.Rows[i]["KategoriSiteMapGorunmesin"].ToString()) == true)
            { urllink = "javascript:;"; }
            else
            { urllink = (islem.GlobalSiteLinkSec(rsn.Rows[i]["SayfaTurID"].ToString(), rsn.Rows[i]["id"].ToString())); }
            //----------------------------------------------------------------------------------------------
            Agac.Add(Fonksiyon.IlkHarflerBuyuk(rsn.Rows[i]["Kategori"].ToString()), urllink);
        }
        Agac.Add(Fonksiyon.IlkHarflerBuyuk(gelenveri), "");
        Agac.RegisterAgacSite(ltrlAgac);
        //----------------------------------------------------------------------------------------
        Fonksiyon.SetTitle(Fonksiyon.IlkHarflerBuyuk(title), sayfa);
        //----------------------------------------------------------------------------------------
    }
    //-------------------------------------------------------------------------------------------

    public static string AgacHrefTitleSec(string gelenurl, string gelenadi)
    {
        //------------------------------------------------------------------------------
        string geridonenveri = null;
        //------------------------------------------------------------------------------
        DataRow rs = Baglanti.GetDataRow("select ResimTitleTag from Tbl_Genel_Kategori where KategoriUrl='" + Fonksiyon.SqlTemizle(AgacHrefGlobalRequestSiz(gelenurl)) + "' ");
        if (rs != null)
        { geridonenveri = rs[0].ToString(); }
        else if (String.IsNullOrEmpty(geridonenveri))
        {
            DataRow rs1 = Baglanti.GetDataRow("select ResimTitleTag from Tbl_Genel_Kayitlar where BaslikUrl='" + Fonksiyon.SqlTemizle(AgacHrefGlobalRequestSiz(gelenurl)) + "' ");
            if (rs1 != null)
            { geridonenveri = rs1[0].ToString(); }
            else
            { geridonenveri = gelenadi; }
        }
        else
        { geridonenveri = gelenadi; }
        //------------------------------------------------------------------------------
        return geridonenveri;
        //------------------------------------------------------------------------------
    }
    //----------------------------------------------------------------------------------------------

    public static string AgacHrefGlobalRequestSiz(string url)
    {
        try
        {
            string katlink = Fonksiyon.SqlInjectionTemizle(url.Replace("/", "").Substring(0, url.Substring(1).IndexOf("/")));
            return katlink;
        }
        catch { return null; }
    }
    //--------------------------------------------------------------------------------------------

    //----------------------------------------------------------------------------------------------
    //--------------------------Agaç İşlemleri Son-------------------------------------------------
    //----------------------------------------------------------------------------------------------


    //----------------------------------------------------------------------------------------------
    //--------------------------Diğer Ek Kisayol İşlemleri Başlangıç-------------------------------
    //----------------------------------------------------------------------------------------------




    //----------------------------------------------------------------------------------------------
    public static string NoIndexDurumSec(string id, int tur)
    {
        //----------------------------------------------------------------------------------------
        string geridonenveri = null;
        //----------------------------------------------------------------------------------------

        //----------------------------------------------------------------------------------------
        DataRow rs = null;
        //----------------------------------------------------------------------------------------

        //----------------------------------------------------------------------------------------
        if (tur == 1)
        {
            rs = Baglanti.GetDataRow("select top 1 id from Tbl_Genel_Kategori where id=" + Convert.ToInt32(id) + " and NoIndexDurum=1 and Onay=1 and SilDurum=0");
            if (rs != null)
            { geridonenveri = "<meta name=\"robots\" content=\"noindex, nofollow\" />"; }
        }
        else if (tur == 2)
        {
            rs = Baglanti.GetDataRow("select top 1 id from Tbl_Genel_Kayitlar where id=" + Convert.ToInt32(id) + " and NoIndexDurum=1 and Onay=1 and SilDurum=0");
            if (rs != null)
            { geridonenveri = "<meta name=\"robots\" content=\"noindex, nofollow\" />"; }
        }
        else if (tur == 3)
        {
            rs = Baglanti.GetDataRow("select top 1 id from Tbl_Zm_SayfaTurleri where id=" + Convert.ToInt32(id) + " and NoIndexDurum=1");
            if (rs != null)
            { geridonenveri = "<meta name=\"robots\" content=\"noindex, nofollow\" />"; }
        }
        //---------------------------------------------------------------------------------
        return geridonenveri;
    }
    //---------------------------------------------------------------------------------------------




    //----------------------------------------------------------------------------------------------
    public static string KategoriBannerResimSec(string id)
    {
        string geridonenveri = null;
        //----------------------------------------------------------------------------------------
        string AtkifID = islem.CokluEnUstKategoriIDSec(Convert.ToInt32(id));
        string AtkifUID = islem.VeriGetirSec(id, "Genel_Kategori", "UstKategoriID", "id");
        //----------------------------------------------------------------------------------------
        if (!String.IsNullOrEmpty(islem.VeriGetirSec(id, "Genel_Kategori", "BannerResim", "id")))
        { id = id; }
        else if (!String.IsNullOrEmpty(islem.VeriGetirSec(AtkifUID, "Genel_Kategori", "BannerResim", "id")))
        { id = AtkifUID; }
        else if (!String.IsNullOrEmpty(islem.VeriGetirSec(AtkifID, "Genel_Kategori", "BannerResim", "id")))
        { id = AtkifID; }
        //----------------------------------------------------------------------------------------
        DataRow rs = Baglanti.GetDataRow("select top 1 BannerResim from Tbl_Genel_Kategori where id=" + Convert.ToInt32(id) + " and BannerResim<>'' and  Onay=1 and SilDurum=0");
        if (rs != null)
        { geridonenveri = rs["BannerResim"].ToString(); }
        //---------------------------------------------------------------------------------
        return geridonenveri;
    }
    //---------------------------------------------------------------------------------------------

    //----------------------------------------------------------------------------------------------
    public static string DilGetirSec(int gelenveri)
    {
        string geridonenveri = null;
        //---------------------------------------------------------
        DataRow DilTabloRS = Baglanti.GetDataRow("select Baslik from Tbl_Diller where id=" + Convert.ToInt32(gelenveri) + "");
        if (DilTabloRS != null)
        { geridonenveri = DilTabloRS[0].ToString(); }
        //---------------------------------------------------------
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string KatalogHrefSec(int id)
    {
        string geridonenveri = null;
        try
        {
            DataRow rs = Baglanti.GetDataRow("select id,TurID,Dosya,Link from Tbl_Genel_Kataloglar where id=" + System.Convert.ToInt32(id) + " and Onay=1 ");
            if (rs != null)
            {
                if (Convert.ToInt32(rs["TurID"].ToString()) == 0)
                { geridonenveri = Fonksiyon.DosyaKlasoru + rs["Dosya"].ToString(); }
                else if (Convert.ToInt32(rs["TurID"].ToString()) == 1)
                { geridonenveri = "/FlashKatalog/Default.aspx?CatID=" + rs["id"].ToString() + ""; }
                else if (Convert.ToInt32(rs["TurID"].ToString()) == 2)
                { geridonenveri = rs["Link"].ToString(); }
            }
        }
        catch { }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string SiralamaSekliSec(int gelenveri)
    {
        string geridonenveri = null;
        //---------------------------------------    
        try
        {
            DataRow rs = Baglanti.GetDataRow("Select Tbl_Zm_SiralamaSekilleri.Siralama From Tbl_Genel_Kategori INNER JOIN Tbl_Zm_SiralamaSekilleri ON Tbl_Genel_Kategori.SiralamaSekli = Tbl_Zm_SiralamaSekilleri.id where Tbl_Genel_Kategori.id='" + Convert.ToInt32(gelenveri) + "'");
            if (rs != null)
            { geridonenveri = rs[0].ToString(); }
            else
            { geridonenveri = " sira asc, id desc"; }
        }
        catch
        { geridonenveri = " sira asc, id desc"; }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string SiralamaSekliFormSec(int gelenveri)
    {
        string geridonenveri = null;
        //---------------------------------------    
        try
        {
            DataRow rs = Baglanti.GetDataRow("Select Tbl_Zm_SiralamaSekilleri.Siralama From Tbl_Genel_Kategori INNER JOIN Tbl_Zm_SiralamaSekilleri ON Tbl_Genel_Kategori.SiralamaSekli = Tbl_Zm_SiralamaSekilleri.id where Tbl_Genel_Kategori.id='" + Convert.ToInt32(gelenveri) + "'");
            if (rs != null)
            { geridonenveri = rs[0].ToString(); }
            else
            { geridonenveri = " sira asc, id asc"; }
        }
        catch
        { geridonenveri = " sira asc, id asc"; }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string MenuSiralamaSekliSec(int gelenveri)
    {
        string geridonenveri = null;
        //---------------------------------------    
        try
        {
            DataRow rs = Baglanti.GetDataRow("Select Tbl_Zm_SiralamaSekilleri.Siralama From Tbl_Genel_Kategori INNER JOIN Tbl_Zm_SiralamaSekilleri ON Tbl_Genel_Kategori.SiralamaSekli = Tbl_Zm_SiralamaSekilleri.id where Tbl_Genel_Kategori.id='" + Convert.ToInt32(gelenveri) + "'");
            if (rs != null)
            { geridonenveri = rs[0].ToString(); }
            else
            { geridonenveri = " sira asc, id asc"; }
        }
        catch
        { geridonenveri = " sira asc, id asc"; }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string CropResimBelirt(string Durum)
    {
        switch ((Durum))
        {
            case "b_": Durum = "( Büyük Resim )"; break;
            case "o_": Durum = "( Orta Resim )"; break;
            case "k_": Durum = "( Küçük Resim )"; break;
        }
        return Durum;
    }
    //----------------------------------------------------------------------------------------------

    public static string BannerReklamSec(int gelenveri)
    {
        System.Text.StringBuilder shtml = new System.Text.StringBuilder();
        DataRow ReklamTabloRS = Baglanti.GetDataRow("select TurID,YerID,Url,AcilisSekli,Resim,Genislik,Yukseklik,ReklamKodu,Baslik from Tbl_Reklamlar where YerID='" + Convert.ToInt32(gelenveri) + "' and Onay=1");
        if (ReklamTabloRS != null)
        {
            if (Convert.ToInt32(ReklamTabloRS["TurID"]) == 0)
            {
                if (!String.IsNullOrEmpty(ReklamTabloRS["Url"].ToString()))
                { shtml.Append("<a href=\"" + ReklamTabloRS["Url"].ToString() + "\" target=\"" + ReklamTabloRS["AcilisSekli"].ToString() + "\"  title=\"" + ReklamTabloRS["Baslik"].ToString() + "\"><img alt=\"" + ReklamTabloRS["Baslik"].ToString() + "\" src=\"" + Fonksiyon.ResimKlasoru + "" + ReklamTabloRS["Resim"].ToString() + "\" width=\"" + ReklamTabloRS["Genislik"].ToString() + "\" height=\"" + ReklamTabloRS["Yukseklik"].ToString() + "\" border=\"0\"></a>"); }
                else
                { shtml.Append("<img alt=\"" + ReklamTabloRS["Baslik"].ToString() + "\" src=\"" + Fonksiyon.ResimKlasoru + "" + ReklamTabloRS["Resim"].ToString() + "\" width=\"" + ReklamTabloRS["Genislik"].ToString() + "\" height=\"" + ReklamTabloRS["Yukseklik"].ToString() + "\" border=\"0\">"); }
            }
            else if (Convert.ToInt32(ReklamTabloRS["TurID"]) == 1)
            {
                shtml.Append("<object classID=\"clsid:D27CDB6E-AE6D-11CF-96B8-444553540000\" ID=\"Object1\" codebase=\"download.macromedia.com/pub/shockwave/cabs/Banners/swflash.cab#version=6,0,40,0\" border=\"0\"  width=\"" + ReklamTabloRS["Genislik"].ToString() + "\" height=\"" + ReklamTabloRS["Yukseklik"].ToString() + "\"  style=\"text-align:center;\">");
                shtml.Append("<param name=\"movie\" value=\"" + Fonksiyon.ResimKlasoru + "" + ReklamTabloRS["Resim"].ToString() + "\" />");
                shtml.Append("<param name=\"quality\" value=\"High\" />");
                shtml.Append("<param name=\"wmode\" value=\"transparent\" />");
                shtml.Append("<embed src=\"" + Fonksiyon.ResimKlasoru + "" + ReklamTabloRS["Resim"].ToString() + "\" pluginspage=\"www.macromedia.com/go/getflashplayer\" type=\"application/x-shockwave-flash\" name=\"Object2\"  width=\"" + ReklamTabloRS["Genislik"].ToString() + "\" height=\"" + ReklamTabloRS["Yukseklik"].ToString() + "\"  quality=\"High\" wmode=\"transparent\" />");
                shtml.Append("</object>");
            }
            else if (Convert.ToInt32(ReklamTabloRS["TurID"]) == 2)
            { shtml.Append(ReklamTabloRS["ReklamKodu"].ToString().Replace("´", "'")); }
        }
        return shtml.ToString();
    }
    //----------------------------------------------------------------------------------------------

    public static string AcilisReklamSec()
    {
        System.Text.StringBuilder shtml = new System.Text.StringBuilder();
        DataRow rs = Baglanti.GetDataRow("select PopupDurum,PopupTur,PopupSuresi,PopupGenislik,PopupYukseklik,PopupUrl,PopupAcilisSekli,PopupResim,Popupicerik from Tbl_SiteAyarlari");
        if (rs != null)
        {
            if (Convert.ToBoolean(rs["PopupDurum"]) == true && Convert.ToInt32(HttpContext.Current.Session["Reklam"]) == 0)
            {
                HttpContext.Current.Session["Reklam"] = 1;
                HttpContext.Current.Response.Redirect("/default.aspx?interstitial=true");
            }

            if (HttpContext.Current.Request["interstitial"] == "true")
            {

                if (Convert.ToInt32(rs["PopupTur"]) == 0)
                {
                    if (!String.IsNullOrEmpty(rs["PopupUrl"].ToString()))
                    {
                        shtml.Append("<a href=\"" + rs["PopupUrl"].ToString() + "\" target=\"" + rs["PopupAcilisSekli"].ToString() + "\"><img src=\"" + Fonksiyon.ResimKlasoru + "" + rs["PopupResim"].ToString() + "\" width=\"" + rs["PopupGenislik"].ToString() + "\" height=\"" + rs["PopupYukseklik"].ToString() + "\" border=\"0\"></a>");
                    }
                    else
                    {
                        shtml.Append("<img src=\"" + Fonksiyon.ResimKlasoru + "" + rs["PopupResim"].ToString() + "\" width=\"" + rs["PopupGenislik"].ToString() + "\" height=\"" + rs["PopupYukseklik"].ToString() + "\" border=\"0\">");
                    }
                }
                else if (Convert.ToInt32(rs["PopupTur"]) == 1)
                {
                    shtml.Append("<object classID=\"clsid:D27CDB6E-AE6D-11CF-96B8-444553540000\" ID=\"Object1\" codebase=\"download.macromedia.com/pub/shockwave/cabs/Banners/swflash.cab#version=6,0,40,0\" border=\"0\"  width=\"" + rs["PopupGenislik"].ToString() + "\" height=\"" + rs["PopupYukseklik"].ToString() + "\"  style=\"text-align:center;\">");
                    shtml.Append("<param name=\"movie\" value=\"" + Fonksiyon.ResimKlasoru + "" + rs["PopupResim"].ToString() + "\" />");
                    shtml.Append("<param name=\"quality\" value=\"High\" />");
                    shtml.Append("<param name=\"wmode\" value=\"transparent\" />");
                    shtml.Append("<embed src=\"" + Fonksiyon.ResimKlasoru + "" + rs["PopupResim"].ToString() + "\" pluginspage=\"www.macromedia.com/go/getflashplayer\" type=\"application/x-shockwave-flash\" name=\"Object2\"  width=\"" + rs["PopupGenislik"].ToString() + "\" height=\"" + rs["PopupYukseklik"].ToString() + "\"  quality=\"High\" wmode=\"transparent\" />");
                    shtml.Append("</object>");
                }
                else if (Convert.ToInt32(rs["PopupTur"]) == 2)
                {
                    shtml.Append(rs["Popupicerik"].ToString().Replace("´", "'"));
                }

            }
        }
        return shtml.ToString();
    }
    //----------------------------------------------------------------------------------------------

    public static bool PingAtmak()
    {
        bool durum = true;
        //----------------------------------------------------------------------------------------
        try
        {
            string pingurl_pingomatic = "http://pingomatic.com/ping/?title=" + Ayarlar.Siteismi + "&blogurl=http://" + HttpContext.Current.Request.ServerVariables["HTTP_HOST"] + "&rssurl=http://" + HttpContext.Current.Request.ServerVariables["HTTP_HOST"] + "/rss.xml&chk_weblogscom=on&chk_blogs=on&chk_technorati=on&chk_feedburner=on&chk_syndic8=on&chk_newsgator=on&chk_myyahoo=on&chk_pubsubcom=on&chk_blogdigger=on&chk_blogrolling=on&chk_blogstreet=on&chk_moreover=on&chk_weblogalot=on&chk_icerocket=on&chk_newsisfree=on&chk_topicexchange=on&chk_google=on&chk_tailrank=on&chk_bloglines=on&chk_aiderss=on";
            HttpWebRequest request_pingomatic = (HttpWebRequest)WebRequest.Create(pingurl_pingomatic);
            request_pingomatic.Method = "POST";
            request_pingomatic.ContentType = "text/xml";
            request_pingomatic.Timeout = 3000;
            request_pingomatic.GetResponse();
        }
        catch
        { durum = false; }
        //----------------------------------------------------------------------------------------
        return durum;
    }
    //--------------------------------------------------------------------------------------------

    public static bool PingTekliAtmak(string CatID, string id, string Baslik, string BaslikUrl)
    {
        bool durum = true;
        //----------------------------------------------------------------------------------------
        try
        {
            string url = GlobalSiteDetayLinkSec(CatID, id);
            //----------------------------------------------------------------------------------------
            string pingurl_pingomatic = "http://pingomatic.com/ping/?title=" + Fonksiyon.XSSTemizle(Baslik) + "&blogurl=http://" + HttpContext.Current.Request.ServerVariables["HTTP_HOST"] + "" + url + "&chk_weblogscom=on&chk_blogs=on&chk_technorati=on&chk_feedburner=on&chk_syndic8=on&chk_newsgator=on&chk_myyahoo=on&chk_pubsubcom=on&chk_blogdigger=on&chk_blogrolling=on&chk_blogstreet=on&chk_moreover=on&chk_weblogalot=on&chk_icerocket=on&chk_newsisfree=on&chk_topicexchange=on&chk_google=on&chk_tailrank=on&chk_bloglines=on&chk_aiderss=on";
            HttpWebRequest request_pingomatic = (HttpWebRequest)WebRequest.Create(pingurl_pingomatic);
            request_pingomatic.Method = "POST";
            request_pingomatic.ContentType = "text/xml";
            request_pingomatic.Timeout = 3000;
            request_pingomatic.GetResponse();
        }
        catch
        { durum = false; }
        //----------------------------------------------------------------------------------------
        return durum;
    }
    //--------------------------------------------------------------------------------------------

    public static string EtiketSec(string gelenveri)
    {
        System.Text.StringBuilder shtml = new System.Text.StringBuilder();
        int i = 0;
        string s = gelenveri;
        string[] words = s.Split(',');
        foreach (string word in words)
        {
            if (i == 0) { shtml.Append("<a href=\"/etiket/" + word.Replace(" ", "+") + ".html\">" + word + "</a>"); }
            else { shtml.Append(" , <a href=\"/etiket/" + word.Replace(" ", "+") + ".html\">" + word + "</a>"); }
            i++;
        }
        return shtml.ToString();
    }
    //----------------------------------------------------------------------------------------------

    public static string HeadBolumSec(string Description, string Keywords, Literal lbldescription, Literal lblkeywords)
    {
        string geridonenveri = null;
        //----------------------------------------------------------------------------------
        if (!String.IsNullOrEmpty(Description))
        { lbldescription.Text = "\n<meta name=\"Description\" content=\"" + Description + "\">\n"; HttpContext.Current.Session["HeadBolumD"] = "Description"; }
        //----------------------------------------------------------------------------------

        if (!String.IsNullOrEmpty(Keywords))
        { lblkeywords.Text = "<meta name=\"keywords\" content=\"" + Keywords + "\">\n"; HttpContext.Current.Session["HeadBolumK"] = "Keywords"; }
        //----------------------------------------------------------------------------------
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    //----------------------------------------------------------------------------------------------
    //--------------------------Diğer Ek Kisayol İşlemleri Son-------------------------------------
    //----------------------------------------------------------------------------------------------


    //--------------------------------------------------------------------------------------------
    //-------------------------------Kategori ve Kayıt İşlemleri Başlangıç -----------------------
    //--------------------------------------------------------------------------------------------

    public static string GlobalEnUstKategoriUrlSec(int UstKategoriID)
    {
        string geridonenveri = "";
        DataRow rs = Baglanti.GetDataRow("Select TOP 1 KategoriUrl,UstKategoriID From Tbl_Genel_Kategori Where id=" + Convert.ToInt32(UstKategoriID) + "");
        if (rs != null)
        {
            geridonenveri = rs["KategoriUrl"].ToString();
            if (rs["UstKategoriID"].ToString() != "0")
            { geridonenveri = GlobalEnUstKategoriUrlSec(Convert.ToInt32(rs["UstKategoriID"].ToString())); }
        }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string GlobalKategoriUrlSec(int UstKategoriID)
    {
        string geridonenveri = "";
        DataRow rs = Baglanti.GetDataRow("select KategoriUrl,UstKategoriID from Tbl_Genel_Kategori where id=" + System.Convert.ToInt32(UstKategoriID) + "");
        if (rs != null)
        {
            geridonenveri = rs["KategoriUrl"].ToString() + "/";
            if (rs["UstKategoriID"].ToString() != "0")
            { geridonenveri = GlobalKategoriUrlSec(Convert.ToInt32(rs["UstKategoriID"].ToString())) + geridonenveri; }
        }
        return geridonenveri;
    }
    //--------------------------------------------------------------------------------------------

    public static string GloblSayfalamaSec(int UstKategoriID)
    {
        //---------------------------------------------------------------------------------------------------
        string geridonenveri = "";
        //---------------------------------------------------------------------------------------------------
        DataRow rs = Baglanti.GetDataRow("select id,UstKategoriID,KategoriUrl,SayfaTurID,DevamiSayfaTurID from Tbl_Genel_Kategori where id=" + Convert.ToInt32(UstKategoriID) + "");
        if (rs != null)
        {
            if (Ayarlar.UrlLinkTur == 0)
            {
                if (GlobalEnUstKategoriUrlSec(Convert.ToInt32(rs["id"].ToString())) == rs["KategoriUrl"].ToString())
                { geridonenveri = "/" + rs["KategoriUrl"].ToString() + "/"; }
                else if (Convert.ToInt32(islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "SayfaTurID", "id")) != Convert.ToInt32(rs["SayfaTurID"].ToString()) && (Convert.ToInt32(islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "DevamiSayfaTurID", "id")) != Convert.ToInt32(rs["SayfaTurID"].ToString())))
                { geridonenveri = "/" + rs["KategoriUrl"].ToString() + "/"; }
                else
                { geridonenveri = "/" + GlobalEnUstKategoriUrlSec(Convert.ToInt32(rs["id"].ToString())) + "/" + rs["KategoriUrl"].ToString() + "/"; }

            }
            else if (Ayarlar.UrlLinkTur == 1)
            {
                if (Convert.ToInt32(islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "SayfaTurID", "id")) != Convert.ToInt32(rs["SayfaTurID"].ToString()))
                { geridonenveri = "/" + rs["KategoriUrl"].ToString() + "/"; }
                else
                { geridonenveri = "/" + islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["KategoriUrl"].ToString() + "/"; }
            }
            else if (Ayarlar.UrlLinkTur == 2)
            { geridonenveri = "/" + GlobalKategoriUrlSec(Convert.ToInt32(UstKategoriID)) + "/" + rs["KategoriUrl"].ToString() + "/"; }
            else if (Ayarlar.UrlLinkTur == 3)
            { geridonenveri = "/" + rs["KategoriUrl"].ToString() + "/"; }
            geridonenveri = geridonenveri.Replace("//", "/");
        }
        //---------------------------------------------------------------------------------------------------
        return geridonenveri;
        //---------------------------------------------------------------------------------------------------
    }
    //-------------------------------------------------------------------------------------------

    public static string GlobalSiteDetayLinkSec(string catid, string id)
    {
        string geridonenveri = null;
        //--------------------------------------------------------
        DataRow rs = Baglanti.GetDataRow("SELECT Tbl_Genel_Kategori.KategoriUrl, Tbl_Genel_Kategori.UstKategoriID, Tbl_Genel_Kategori.SayfaTurID, Tbl_Zm_SayfaTurleri.SiteMapGorunmesin, Tbl_Genel_Kategori.MenuDetayTekliDurum,Tbl_Zm_SayfaTurleri.DetayTekliDurum, Tbl_Genel_Kayitlar.BaslikUrl, Tbl_Genel_Kayitlar.Link FROM Tbl_Genel_Kategori INNER JOIN Tbl_Genel_Kayitlar ON Tbl_Genel_Kategori.id = Tbl_Genel_Kayitlar.CatID INNER JOIN Tbl_Zm_SayfaTurleri ON Tbl_Genel_Kategori.SayfaTurID = Tbl_Zm_SayfaTurleri.id WHERE Tbl_Genel_Kayitlar.CatID=" + Convert.ToInt32(catid) + " and Tbl_Genel_Kayitlar.id=" + Convert.ToInt32(id) + " and Tbl_Genel_Kayitlar.Onay=1 and Tbl_Genel_Kayitlar.SilDurum=0 ");
        if (rs != null)
        {
            if (Convert.ToBoolean(rs["SiteMapGorunmesin"].ToString()) == false)
            {
                if (String.IsNullOrEmpty(rs["Link"].ToString()))
                {

                    if (Convert.ToBoolean(rs["DetayTekliDurum"].ToString()) == true || CokluEnUstMenuDetayTekliDurumSec(Convert.ToInt32(catid)) == true)
                    { geridonenveri = "/" + rs["BaslikUrl"].ToString() + ".html"; }
                    else
                    {
                        //-------------------------------------------------------------------------------------------------------------------------
                        if (Ayarlar.UrlLinkTur == 0)
                        { geridonenveri = "/" + islem.GlobalEnUstKategoriUrlSec(Convert.ToInt32(catid)) + "/" + rs["BaslikUrl"].ToString() + ".html"; }
                        else if (Ayarlar.UrlLinkTur == 1)
                        {
                            //--------------------------------------------------------------------------
                            if (rs["KategoriUrl"].ToString() == rs["BaslikUrl"].ToString())
                            { geridonenveri = "/" + islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["BaslikUrl"].ToString() + ".html"; }
                            else
                            { geridonenveri = "/" + islem.VeriGetirSec(catid, "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["BaslikUrl"].ToString() + ".html"; }
                            //--------------------------------------------------------------------------
                        }
                        else if (Ayarlar.UrlLinkTur == 2)
                        { geridonenveri = "/" + islem.GlobalKategoriUrlSec(Convert.ToInt32(catid)) + "/" + rs["BaslikUrl"].ToString() + ".html"; }
                        else if (Ayarlar.UrlLinkTur == 3)
                        {
                            //--------------------------------------------------------------------------
                            //geridonenveri = "/" + rs["BaslikUrl"].ToString() + ".html"; 
                            //--------------------------------------------------------------------------
                            if (rs["KategoriUrl"].ToString() == rs["BaslikUrl"].ToString())
                            { geridonenveri = "/" + islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["BaslikUrl"].ToString() + ".html"; }
                            else
                            { geridonenveri = "/" + islem.VeriGetirSec(catid, "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["BaslikUrl"].ToString() + ".html"; }
                            //--------------------------------------------------------------------------
                        }
                        //-------------------------------------------------------------------------------------------------------------------------
                    }
                }
                else
                { geridonenveri = rs["Link"].ToString(); }
                //----------------------------------------

                if (geridonenveri == null)
                {
                    if (Ayarlar.GlobalAsaxHtmlSizDurum == true)
                    { return geridonenveri.Replace(".html", "/"); }
                    else
                    { return geridonenveri ; }
                }
                else
                {
                    if (Ayarlar.GlobalAsaxHtmlSizDurum == true)
                    { return geridonenveri.Replace("//", "/").Replace(".html", "/");  }
                    else
                    { return geridonenveri.Replace("//", "/"); }
                }
                //----------------------------------------
            }
            else
            { 
                return geridonenveri = GlobalSiteLinkSec(rs["SayfaTurID"].ToString(), catid.ToString()); 
            }
            //----------------------------------------
        }
        else
        { return geridonenveri = null; }
        //----------------------------------------------------------------------------------------------
    }
    //--------------------------------------------------------------------------------------------

    public static string GlobalSiteDetayEkLinkSec(string catid, string id, string tabloadi, string kayiturl)
    {
        string geridonenveri = null;
        tabloadi = "Tbl_" + Fonksiyon.SqlInjectionTemizle(tabloadi);
        kayiturl = Fonksiyon.SqlInjectionTemizle(kayiturl);
        //--------------------------------------------------------
        DataRow rs = Baglanti.GetDataRow("SELECT Tbl_Genel_Kategori.KategoriUrl, Tbl_Genel_Kategori.UstKategoriID, Tbl_Genel_Kategori.SayfaTurID, Tbl_Zm_SayfaTurleri.SiteMapGorunmesin, Tbl_Genel_Kategori.MenuDetayTekliDurum,Tbl_Zm_SayfaTurleri.DetayTekliDurum, " + tabloadi + "." + kayiturl + " FROM Tbl_Genel_Kategori INNER JOIN " + tabloadi + " ON Tbl_Genel_Kategori.id = " + tabloadi + ".CatID INNER JOIN Tbl_Zm_SayfaTurleri ON Tbl_Genel_Kategori.SayfaTurID = Tbl_Zm_SayfaTurleri.id WHERE " + tabloadi + ".CatID=" + Convert.ToInt32(catid) + " and " + tabloadi + ".id=" + Convert.ToInt32(id) + " and " + tabloadi + ".Onay=1  ");
        if (rs != null)
        {
            if (Convert.ToBoolean(rs["SiteMapGorunmesin"].ToString()) == false)
            {


                if (Convert.ToBoolean(rs["DetayTekliDurum"].ToString()) == true || CokluEnUstMenuDetayTekliDurumSec(Convert.ToInt32(catid)) == true)
                { geridonenveri = "/" + rs["" + kayiturl + ""].ToString() + ".html"; }
                else
                {
                    //-------------------------------------------------------------------------------------------------------------------------
                    if (Ayarlar.UrlLinkTur == 0)
                    { geridonenveri = "/" + islem.GlobalEnUstKategoriUrlSec(Convert.ToInt32(catid)) + "/" + rs["" + kayiturl + ""].ToString() + ".html"; }
                    else if (Ayarlar.UrlLinkTur == 1)
                    {
                        //--------------------------------------------------------------------------
                        if (rs["KategoriUrl"].ToString() == rs["" + kayiturl + ""].ToString())
                        { geridonenveri = "/" + islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["" + kayiturl + ""].ToString() + ".html"; }
                        else
                        { geridonenveri = "/" + islem.VeriGetirSec(catid, "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["" + kayiturl + ""].ToString() + ".html"; }
                        //--------------------------------------------------------------------------
                    }
                    else if (Ayarlar.UrlLinkTur == 2)
                    { geridonenveri = "/" + islem.GlobalKategoriUrlSec(Convert.ToInt32(catid)) + "/" + rs["" + kayiturl + ""].ToString() + ".html"; }
                    else if (Ayarlar.UrlLinkTur == 3)
                    {
                        //--------------------------------------------------------------------------
                        //geridonenveri = "/" + rs[""+ kayiturl +""].ToString() + ".html"; 
                        //--------------------------------------------------------------------------
                        if (rs["KategoriUrl"].ToString() == rs["" + kayiturl + ""].ToString())
                        { geridonenveri = "/" + islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["" + kayiturl + ""].ToString() + ".html"; }
                        else
                        { geridonenveri = "/" + islem.VeriGetirSec(catid, "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["" + kayiturl + ""].ToString() + ".html"; }
                        //--------------------------------------------------------------------------
                    }
                    //-------------------------------------------------------------------------------------------------------------------------

                }
                //----------------------------------------

                if (geridonenveri == null)
                { return geridonenveri; }
                else
                { return geridonenveri.Replace("//", "/"); }
                //----------------------------------------
            }
            else
            { return geridonenveri = GlobalSiteDetayEkLinkSec(rs["SayfaTurID"].ToString(), catid.ToString(), tabloadi, kayiturl); }
            //----------------------------------------
        }
        else
        { return geridonenveri = null; }
        //----------------------------------------------------------------------------------------------
    }
    //--------------------------------------------------------------------------------------------

    public static string GlobalSabitLinkSec(int gelenveri)
    {
        string geridonenveri = null;
        //--------------------------------------------------------------------------------------------------
        DataRow rs = Baglanti.GetDataRow("SELECT id,SayfaTurID FROM Tbl_Genel_Kategori where id=" + Convert.ToInt32(gelenveri) + " and Onay=1 and SilDurum=0 ");
        if (rs != null)
        {
            geridonenveri = GlobalSiteLinkSec(rs["SayfaTurID"].ToString(), rs["id"].ToString());
            return geridonenveri;
        }
        else
        { return geridonenveri = null; }
    }
    //----------------------------------------------------------------------------------------------

    public static string DevamiGlobalSabitLinkSec(int gelenveri)
    {
        string geridonenveri = null;
        //--------------------------------------------------------------------------------------------------
        DataRow rs = Baglanti.GetDataRow("SELECT id,DevamiSayfaTurID FROM Tbl_Genel_Kategori where id=" + Convert.ToInt32(gelenveri) + " and Onay=1 and SilDurum=0 ");
        if (rs != null)
        {
            geridonenveri = GlobalSiteLinkSec(rs["DevamiSayfaTurID"].ToString(), rs["id"].ToString());
            return geridonenveri;
        }
        else
        { return geridonenveri = null; }
    }
    //----------------------------------------------------------------------------------------------

    public static string GlobalSiteLinkSec(string sayfaturid, string id)
    {
        string geridonenveri = null;
        DataRow rs = Baglanti.GetDataRow("select SiteKategoriUrl,LinksizMenuIlkKayit,LinksizMenuIlkKategori,DetayTekliDurum,DirekListele from Tbl_Zm_SayfaTurleri where id=" + Convert.ToInt32(sayfaturid) + "");
        if (rs != null)
        {
            string klink = "";
            //--------------------------------------------------------------------------------------------------
            DataRow rsk = Baglanti.GetDataRow("select id,UstKategoriID,SayfaTurID,Link,GorunumSekli,DevamiSayfaTurID,KategoriUrl,SiteLink,KategoriSiteMapGorunmesin,KategoriDirekListele from Tbl_Genel_Kategori where id=" + Convert.ToInt32(id) + "");
            if (rsk != null)
            {
                if (Convert.ToBoolean(rsk["KategoriSiteMapGorunmesin"].ToString()) == true)
                { geridonenveri = "javascript:;"; }
                else if (String.IsNullOrEmpty(rsk["Link"].ToString()))
                {
                    if (sayfaturid != "1")
                    {
                        //---------------------------------------------------------------------
                        if (Convert.ToBoolean(rs["LinksizMenuIlkKayit"]) == true)
                        { geridonenveri = GlobalUrlIlkKayitBelirtSec(rsk["id"].ToString(), rsk["KategoriUrl"].ToString(), rsk["DevamiSayfaTurID"].ToString()); }
                        else if (Convert.ToBoolean(rs["LinksizMenuIlkKategori"]) == true)
                        { geridonenveri = GlobalUrlIlkKategoriBelirtSec(rsk["id"].ToString(), rsk["KategoriUrl"].ToString(), rsk["DevamiSayfaTurID"].ToString(), Convert.ToBoolean(rs["DirekListele"].ToString())); }
                        else
                        { geridonenveri = GlobalUrlBelirtSec(rsk["id"].ToString(), rsk["UstKategoriID"].ToString(), rsk["KategoriUrl"].ToString(), rsk["SayfaTurID"].ToString(), rs["SiteKategoriUrl"].ToString(), Convert.ToBoolean(rs["DetayTekliDurum"].ToString()), Convert.ToBoolean(rsk["KategoriDirekListele"].ToString()), Convert.ToBoolean(rs["DirekListele"].ToString())); }
                        //----------------------------------------
                        if (geridonenveri == null)
                        { return geridonenveri; }
                        else
                        {
                            if (Ayarlar.GlobalAsaxHtmlSizDurum == true)
                            { return geridonenveri.Replace("//", "/").Replace(".html", "/"); }
                            else
                            { return geridonenveri.Replace("//", "/"); }
                        }
                        //----------------------------------------
                    }
                    else
                    { geridonenveri = "/"; }
                }
                else
                {
                    geridonenveri = rsk["Link"].ToString();
                    klink = rsk["Link"].ToString();
                }
            }
            //----------------------------------------
            if (Ayarlar.GlobalAsaxHtmlSizDurum == true && String.IsNullOrEmpty(klink))
            { return geridonenveri.Replace(".html","/") ; }
            else
            { return geridonenveri; }
            //----------------------------------------

        }
        else
        { return geridonenveri = null; }
    }
    //--------------------------------------------------------------------------------------------

    public static string GlobalUrlBelirtSec(string id, string UstKategoriID, string KategoriUrl, string SayfaTurID, string SiteKategoriUrl, bool DetayTekliDurum, bool KategoriDirekListele, bool SayfaDirekListele)
    {
        string geridonenveri = null;
        //--------------------------------------------------------
        if (Convert.ToInt32(UstKategoriID) == 0)
        {
            if (SiteKategoriUrl.ToString().IndexOf("#") == 0 || SiteKategoriUrl.ToString().IndexOf("javascript:;") == 0)
            { geridonenveri = SiteKategoriUrl; }
            else
            {
                if (Convert.ToBoolean(SayfaDirekListele) == false)
                {
                    if (Convert.ToBoolean(KategoriDirekListele) == false)
                    {
                        DataTable rsyv = Baglanti.GetDataTable("SELECT top 2 id,BaslikUrl  FROM Tbl_Genel_Kayitlar where CatID=" + Convert.ToString(id) + " and Onay=1 and SilDurum=0");
                        if (rsyv.Rows.Count == 1)
                        {
                            if (Convert.ToBoolean(DetayTekliDurum) == true)
                            { geridonenveri = "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                            else
                            { 
                                //geridonenveri = "/" + KategoriUrl + "/"; 
                                geridonenveri = "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + "/";
                            }
                        }
                        else
                        { geridonenveri = "/" + KategoriUrl + "/"; }
                    }
                    else
                    { geridonenveri = "/" + KategoriUrl + "/"; }
                }
                else
                { geridonenveri = "/" + KategoriUrl + "/"; }
            }
        }
        else
        {

            DataTable rsyv = Baglanti.GetDataTable("SELECT top 2 Tbl_Genel_Kayitlar.CatID,Tbl_Zm_SayfaTurleri.DetayTekliDurum,Tbl_Genel_Kayitlar.BaslikUrl, Tbl_Genel_Kayitlar.Link, Tbl_Genel_Kayitlar.BaslikUrl FROM Tbl_Genel_Kategori INNER JOIN Tbl_Genel_Kayitlar ON Tbl_Genel_Kategori.id = Tbl_Genel_Kayitlar.CatID INNER JOIN Tbl_Zm_SayfaTurleri ON Tbl_Genel_Kategori.SayfaTurID = Tbl_Zm_SayfaTurleri.id WHERE Tbl_Genel_Kayitlar.CatID in (" + CokluKategoriIDSec(id) + ") and Tbl_Genel_Kayitlar.Onay=1 and Tbl_Genel_Kayitlar.SilDurum=0");
            if (rsyv.Rows.Count == 1 && Convert.ToBoolean(SayfaDirekListele) == false)
            {
                if (Convert.ToBoolean(KategoriDirekListele) == false)
                {
                    if (String.IsNullOrEmpty(rsyv.Rows[0]["Link"].ToString()))
                    {
                        if (Convert.ToBoolean(rsyv.Rows[0]["DetayTekliDurum"].ToString()) == true || CokluEnUstMenuDetayTekliDurumSec(Convert.ToInt32(rsyv.Rows[0]["CatID"].ToString())) == true)
                        { geridonenveri = "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                        else
                        {
                            //--------------------------------------------------------------------------
                            if (Ayarlar.UrlLinkTur == 0)
                            { geridonenveri = "/" + islem.GlobalEnUstKategoriUrlSec(Convert.ToInt32(id)) + "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                            else if (Ayarlar.UrlLinkTur == 1)
                            {
                                //--------------------------------------------------------------------------
                                if (KategoriUrl == rsyv.Rows[0]["BaslikUrl"].ToString())
                                { geridonenveri = "/" + islem.VeriGetirSec(UstKategoriID, "Genel_Kategori", "KategoriUrl", "id") + "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                                else
                                { geridonenveri = "/" + islem.VeriGetirSec(id, "Genel_Kategori", "KategoriUrl", "id") + "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                                //--------------------------------------------------------------------------
                            }
                            else if (Ayarlar.UrlLinkTur == 2)
                            { geridonenveri = "/" + islem.GlobalKategoriUrlSec(Convert.ToInt32(id)) + "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                            else if (Ayarlar.UrlLinkTur == 3)
                            {
                                //--------------------------------------------------------------------------
                                //geridonenveri = "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; 
                                //--------------------------------------------------------------------------
                                if (KategoriUrl == rsyv.Rows[0]["BaslikUrl"].ToString())
                                { geridonenveri = "/" + islem.VeriGetirSec(UstKategoriID, "Genel_Kategori", "KategoriUrl", "id") + "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                                else
                                { geridonenveri = "/" + islem.VeriGetirSec(id, "Genel_Kategori", "KategoriUrl", "id") + "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                                //--------------------------------------------------------------------------
                            }
                            //--------------------------------------------------------------------------
                        }
                    }
                    else
                    { geridonenveri = rsyv.Rows[0]["Link"].ToString(); }
                }
                else
                {
                    //---------------------------------------------------------------------------------------------------
                    //---------------------------------------------------------------------------------------------------
                    if (Ayarlar.UrlLinkTur == 0)
                    { geridonenveri = "/" + GlobalEnUstKategoriUrlSec(Convert.ToInt32(id)) + "/" + KategoriUrl + "/"; }
                    else if (Ayarlar.UrlLinkTur == 1)
                    {
                        geridonenveri = "/" + islem.VeriGetirSec(UstKategoriID.ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + KategoriUrl + "/";
                    }
                    else if (Ayarlar.UrlLinkTur == 2)
                    { geridonenveri = "/" + GlobalKategoriUrlSec(Convert.ToInt32(UstKategoriID)) + "/" + KategoriUrl + "/"; }
                    else if (Ayarlar.UrlLinkTur == 3)
                    { geridonenveri = "/" + KategoriUrl + "/"; }
                    //---------------------------------------------------------------------------------------------------
                    //---------------------------------------------------------------------------------------------------
                }
            }
            else
            {
                string AnaSayfaTurID = islem.VeriGetirSec(UstKategoriID, "Genel_Kategori", "SayfaTurID", "id");
                //----------------------------------------------------------------------------------------------
                if (Convert.ToInt32(AnaSayfaTurID) == Convert.ToInt32(SayfaTurID))
                {
                    if (SiteKategoriUrl.ToString().IndexOf("#") == 0 || SiteKategoriUrl.ToString().IndexOf("javascript:;") == 0)
                    { geridonenveri = SiteKategoriUrl; }
                    else
                    {
                        //---------------------------------------------------------------------------------------------------
                        //---------------------------------------------------------------------------------------------------
                        if (Ayarlar.UrlLinkTur == 0)
                        { geridonenveri = "/" + GlobalEnUstKategoriUrlSec(Convert.ToInt32(id)) + "/" + KategoriUrl + "/"; }
                        else if (Ayarlar.UrlLinkTur == 1)
                        {
                            geridonenveri = "/" + islem.VeriGetirSec(UstKategoriID.ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + KategoriUrl + "/";
                        }
                        else if (Ayarlar.UrlLinkTur == 2)
                        { geridonenveri = "/" + GlobalKategoriUrlSec(Convert.ToInt32(UstKategoriID)) + "/" + KategoriUrl + "/"; }
                        else if (Ayarlar.UrlLinkTur == 3)
                        { geridonenveri = "/" + KategoriUrl + "/"; }
                        //---------------------------------------------------------------------------------------------------
                        //---------------------------------------------------------------------------------------------------
                    }
                }
                else
                {
                    if (Convert.ToInt32(islem.VeriGetirSec(UstKategoriID, "Genel_Kategori", "DevamiSayfaTurID", "id")) == Convert.ToInt32(SayfaTurID))
                    {
                        if (SiteKategoriUrl.ToString().IndexOf("#") == 0 || SiteKategoriUrl.ToString().IndexOf("javascript:;") == 0)
                        { geridonenveri = SiteKategoriUrl; }
                        else
                        {
                            //---------------------------------------------------------------------------------------------------
                            //---------------------------------------------------------------------------------------------------
                            if (Ayarlar.UrlLinkTur == 0)
                            { geridonenveri = "/" + GlobalEnUstKategoriUrlSec(Convert.ToInt32(id)) + "/" + KategoriUrl + "/"; }
                            else if (Ayarlar.UrlLinkTur == 1)
                            {
                                if (Convert.ToInt32(AnaSayfaTurID) > 5)
                                { geridonenveri = "/" + islem.VeriGetirSec(UstKategoriID.ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + KategoriUrl + "/"; }
                                else
                                { geridonenveri = "/" + KategoriUrl + "/"; }
                            }
                            else if (Ayarlar.UrlLinkTur == 2)
                            { geridonenveri = "/" + GlobalKategoriUrlSec(Convert.ToInt32(UstKategoriID)) + "/" + KategoriUrl + "/"; }
                            else if (Ayarlar.UrlLinkTur == 3)
                            { geridonenveri = "/" + KategoriUrl + "/"; }
                            //---------------------------------------------------------------------------------------------------
                            //---------------------------------------------------------------------------------------------------
                        }
                    }
                    else
                    {
                        string AnaSiteKategoriUrl = islem.VeriGetirSec(AnaSayfaTurID, "Zm_SayfaTurleri", "SiteKategoriUrl", "id");
                        //-------------------------------------------------------------------------------------------------------
                        if (SiteKategoriUrl.ToString().IndexOf("#") == 0 || SiteKategoriUrl.ToString().IndexOf("javascript:;") == 0)
                        { geridonenveri = SiteKategoriUrl; }
                        else
                        { geridonenveri = SiteKategoriUrl + KategoriUrl + "/"; }
                    }
                }
                //----------------------------------------------------------------------------------------------
            }
            //----------------------------------------------------------------------------------------------
        }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string GlobalUrlIlkKategoriBelirtSec(string UstKategoriID, string AnaKategoriUrl, string DevamiSayfaTurID, bool SayfaDirekListele)
    {
        string geridonenveri = null;
        DataRow rs = Baglanti.GetDataRow("select top 1 id,KategoriUrl,KategoriDirekListele,UstKategoriID,SayfaTurID from Tbl_Genel_Kategori where UstKategoriID=" + Convert.ToInt32(UstKategoriID) + " and Onay=1 order by Sira asc, id asc");
        if (rs != null)
        {
            if (Convert.ToBoolean(SayfaDirekListele) == false && Convert.ToBoolean(islem.VeriGetirSec(DevamiSayfaTurID, "Zm_SayfaTurleri", "DirekListele", "id")) == false)
            {
                DataTable rsyv = Baglanti.GetDataTable("SELECT top 2 Tbl_Genel_Kayitlar.CatID,Tbl_Zm_SayfaTurleri.DetayTekliDurum,Tbl_Genel_Kayitlar.BaslikUrl, Tbl_Genel_Kayitlar.Link, Tbl_Genel_Kayitlar.BaslikUrl FROM Tbl_Genel_Kategori INNER JOIN Tbl_Genel_Kayitlar ON Tbl_Genel_Kategori.id = Tbl_Genel_Kayitlar.CatID INNER JOIN Tbl_Zm_SayfaTurleri ON Tbl_Genel_Kategori.SayfaTurID = Tbl_Zm_SayfaTurleri.id WHERE Tbl_Genel_Kayitlar.CatID in (" + CokluKategoriIDSec(rs["id"].ToString()) + ") and Tbl_Genel_Kayitlar.Onay=1 and Tbl_Genel_Kayitlar.SilDurum=0");
                if (rsyv.Rows.Count == 1 && Convert.ToBoolean(rs["KategoriDirekListele"].ToString()) == false)
                {
                    if (String.IsNullOrEmpty(rsyv.Rows[0]["Link"].ToString()))
                    {
                        if (Convert.ToBoolean(rsyv.Rows[0]["DetayTekliDurum"].ToString()) == true || CokluEnUstMenuDetayTekliDurumSec(Convert.ToInt32(rsyv.Rows[0]["CatID"].ToString())) == true)
                        { geridonenveri = "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                        else
                        {
                            //-------------------------------------------------------------------------------------------------------------------------
                            if (Ayarlar.UrlLinkTur == 0)
                            { geridonenveri = "/" + islem.GlobalEnUstKategoriUrlSec(Convert.ToInt32(rs["id"].ToString())) + "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                            else if (Ayarlar.UrlLinkTur == 1)
                            {
                                //------------------------------------------------------------------------------
                                if (rs["KategoriUrl"].ToString() == rsyv.Rows[0]["BaslikUrl"].ToString())
                                { geridonenveri = "/" + islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                                else
                                { geridonenveri = "/" + islem.VeriGetirSec(rs["id"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                                //------------------------------------------------------------------------------
                            }
                            else if (Ayarlar.UrlLinkTur == 2)
                            { geridonenveri = "/" + islem.GlobalKategoriUrlSec(Convert.ToInt32(rs["id"].ToString())) + "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                            else if (Ayarlar.UrlLinkTur == 3)
                            {
                                //------------------------------------------------------------------------
                                //geridonenveri = "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; 
                                //------------------------------------------------------------------------------
                                if (rs["KategoriUrl"].ToString() == rsyv.Rows[0]["BaslikUrl"].ToString())
                                { geridonenveri = "/" + islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                                else
                                { geridonenveri = "/" + islem.VeriGetirSec(rs["id"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rsyv.Rows[0]["BaslikUrl"].ToString() + ".html"; }
                                //------------------------------------------------------------------------------
                            }
                            //-------------------------------------------------------------------------------------------------------------------------
                        }
                    }
                    else
                    { geridonenveri = rsyv.Rows[0]["Link"].ToString(); }
                }
                else
                {
                    //---------------------------------------------------------------------------------------------------
                    //---------------------------------------------------------------------------------------------------
                    if (Ayarlar.UrlLinkTur == 0)
                    {
                        if (GlobalEnUstKategoriUrlSec(Convert.ToInt32(rs["id"].ToString())) == rs["KategoriUrl"].ToString())
                        { geridonenveri = "/" + rs["KategoriUrl"].ToString() + "/"; }
                        else if (Convert.ToInt32(islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "SayfaTurID", "id")) != Convert.ToInt32(rs["SayfaTurID"].ToString()) && (Convert.ToInt32(islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "DevamiSayfaTurID", "id")) != Convert.ToInt32(rs["SayfaTurID"].ToString())))
                        { geridonenveri = "/" + rs["KategoriUrl"].ToString() + "/"; }
                        else
                        { geridonenveri = "/" + GlobalEnUstKategoriUrlSec(Convert.ToInt32(rs["id"].ToString())) + "/" + rs["KategoriUrl"].ToString() + "/"; }

                    }
                    else if (Ayarlar.UrlLinkTur == 1)
                    {
                        if (Convert.ToInt32(islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "SayfaTurID", "id")) != Convert.ToInt32(rs["SayfaTurID"].ToString()))
                        { geridonenveri = "/" + rs["KategoriUrl"].ToString() + "/"; }
                        else
                        {
                            //geridonenveri = "/" + islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["KategoriUrl"].ToString() + "/"; 
                            geridonenveri = "/" + AnaKategoriUrl + "/" + rs["KategoriUrl"].ToString() + "/";
                        }
                    }
                    else if (Ayarlar.UrlLinkTur == 2)
                    { geridonenveri = "/" + GlobalKategoriUrlSec(Convert.ToInt32(UstKategoriID)) + "/" + rs["KategoriUrl"].ToString() + "/"; }
                    else if (Ayarlar.UrlLinkTur == 3)
                    { geridonenveri = "/" + rs["KategoriUrl"].ToString() + "/"; }
                    //---------------------------------------------------------------------------------------------------
                    //---------------------------------------------------------------------------------------------------
                }
            }
            else
            {
                //---------------------------------------------------------------------------------------------------
                //---------------------------------------------------------------------------------------------------
                if (Ayarlar.UrlLinkTur == 0)
                {
                    if (GlobalEnUstKategoriUrlSec(Convert.ToInt32(rs["id"].ToString())) == rs["KategoriUrl"].ToString())
                    { geridonenveri = "/" + rs["KategoriUrl"].ToString() + "/"; }
                    else if (Convert.ToInt32(islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "SayfaTurID", "id")) != Convert.ToInt32(rs["SayfaTurID"].ToString()) && (Convert.ToInt32(islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "DevamiSayfaTurID", "id")) != Convert.ToInt32(rs["SayfaTurID"].ToString())))
                    { geridonenveri = "/" + rs["KategoriUrl"].ToString() + "/"; }
                    else
                    { geridonenveri = "/" + GlobalEnUstKategoriUrlSec(Convert.ToInt32(rs["id"].ToString())) + "/" + rs["KategoriUrl"].ToString() + "/"; }

                }
                else if (Ayarlar.UrlLinkTur == 1)
                {
                    if (Convert.ToInt32(islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "SayfaTurID", "id")) != Convert.ToInt32(rs["SayfaTurID"].ToString()))
                    { geridonenveri = "/" + rs["KategoriUrl"].ToString() + "/"; }
                    else
                    {
                        //geridonenveri = "/" + islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["KategoriUrl"].ToString() + "/"; 
                        geridonenveri = "/" + AnaKategoriUrl + "/" + rs["KategoriUrl"].ToString() + "/";
                    }
                }
                else if (Ayarlar.UrlLinkTur == 2)
                { geridonenveri = "/" + GlobalKategoriUrlSec(Convert.ToInt32(UstKategoriID)) + "/" + rs["KategoriUrl"].ToString() + "/"; }
                else if (Ayarlar.UrlLinkTur == 3)
                { geridonenveri = "/" + rs["KategoriUrl"].ToString() + "/"; }
                //---------------------------------------------------------------------------------------------------
                //---------------------------------------------------------------------------------------------------
            }
        }
        //----------------------------------------------------------------------------------------------
        return geridonenveri;
    }
    //--------------------------------------------------------------------------------------------

    public static string GlobalUrlIlkKayitBelirtSec(string UstKategoriID, string AnaKategoriUrl, string DevamiSayfaTurID)
    {
        string geridonenveri = null;
        //--------------------------------------------------------
        DataRow rs = Baglanti.GetDataRow("SELECT Tbl_Genel_Kategori.KategoriUrl,Tbl_Genel_Kategori.UstKategoriID, Tbl_Genel_Kayitlar.CatID,Tbl_Zm_SayfaTurleri.DetayTekliDurum, Tbl_Genel_Kayitlar.BaslikUrl, Tbl_Genel_Kayitlar.Link, Tbl_Genel_Kayitlar.BaslikUrl, Tbl_Genel_Kayitlar.Sira, Tbl_Genel_Kayitlar.id FROM Tbl_Genel_Kategori INNER JOIN Tbl_Genel_Kayitlar ON Tbl_Genel_Kategori.id = Tbl_Genel_Kayitlar.CatID INNER JOIN Tbl_Zm_SayfaTurleri ON Tbl_Genel_Kategori.SayfaTurID = Tbl_Zm_SayfaTurleri.id WHERE Tbl_Genel_Kayitlar.CatID in (" + CokluKategoriIDSec(UstKategoriID) + ") and Tbl_Genel_Kayitlar.Onay=1 and Tbl_Genel_Kayitlar.SilDurum=0 order by " + MenuSiralamaSekliSec(Convert.ToInt32(UstKategoriID)) + "");
        if (rs != null)
        {
            if (String.IsNullOrEmpty(rs["Link"].ToString()))
            {
                if (Convert.ToBoolean(rs["DetayTekliDurum"].ToString()) == true || CokluEnUstMenuDetayTekliDurumSec(Convert.ToInt32(rs["CatID"].ToString())) == true)
                { geridonenveri = "/" + rs["BaslikUrl"].ToString() + ".html"; }
                else
                {
                    //-------------------------------------------------------------------------------------------------------------------------
                    if (Ayarlar.UrlLinkTur == 0)
                    { geridonenveri = "/" + islem.GlobalEnUstKategoriUrlSec(Convert.ToInt32(rs["CatID"].ToString())) + "/" + rs["BaslikUrl"].ToString() + ".html"; }
                    else if (Ayarlar.UrlLinkTur == 1)
                    {
                        //------------------------------------------------------------------------------
                        if (rs["KategoriUrl"].ToString() == rs["BaslikUrl"].ToString())
                        { geridonenveri = "/" + islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["BaslikUrl"].ToString() + ".html"; }
                        else
                        { geridonenveri = "/" + islem.VeriGetirSec(rs["CatID"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["BaslikUrl"].ToString() + ".html"; }
                        //------------------------------------------------------------------------------
                    }
                    else if (Ayarlar.UrlLinkTur == 2)
                    { geridonenveri = "/" + islem.GlobalKategoriUrlSec(Convert.ToInt32(rs["CatID"].ToString())) + "/" + rs["BaslikUrl"].ToString() + ".html"; }
                    else if (Ayarlar.UrlLinkTur == 3)
                    {
                        //----------------------------------------------------------------
                        //geridonenveri = "/" + rs["BaslikUrl"].ToString() + ".html"; 
                        //----------------------------------------------------------------
                        //------------------------------------------------------------------------------
                        if (rs["KategoriUrl"].ToString() == rs["BaslikUrl"].ToString())
                        { geridonenveri = "/" + islem.VeriGetirSec(rs["UstKategoriID"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["BaslikUrl"].ToString() + ".html"; }
                        else
                        { geridonenveri = "/" + islem.VeriGetirSec(rs["CatID"].ToString(), "Genel_Kategori", "KategoriUrl", "id") + "/" + rs["BaslikUrl"].ToString() + ".html"; }
                        //------------------------------------------------------------------------------
                    }
                    //-------------------------------------------------------------------------------------------------------------------------
                }
            }
            else
            { geridonenveri = rs["Link"].ToString(); }
        }
        //----------------------------------------------------------------------------------------------
        return geridonenveri;
    }
    //-------------------------------------------------------------------------------------------


    //----------------------------------------------------------------------------------------------
    //-------------------------------Kategori ve Kayıt İşlemleri Son-------------------------------
    //----------------------------------------------------------------------------------------------

    //----------------------------------------------------------------------------------------------
    //--------------------------Sayfalama İşlemleri Başlangıç--------------------------------------
    //----------------------------------------------------------------------------------------------

    public static String sayfalamayaphtmlsiz(string sql, int SayfadakiKayitSayisi, int PageID, string yol)
    {
        int GosterilenSayfaSayisi = 2;
        int ToplamKayitSayisi = 0;
        int SayfaBaslangicNumarasi = 0;
        int SayfaBitisNumarasi = 0;
        int SayfaSayisi;
        //------------------------------------------------------------------
        DataTable rs = Baglanti.GetDataTable(sql);
        ToplamKayitSayisi = Convert.ToInt32(rs.Rows.Count);
        SayfaSayisi = Convert.ToInt32(Math.Ceiling(Convert.ToDouble(ToplamKayitSayisi) / Convert.ToDouble(SayfadakiKayitSayisi)));
        //------------------------------------------------------------------
        System.Text.StringBuilder shtml = new System.Text.StringBuilder();
        if (SayfaSayisi > 1)
        {

            SayfaBaslangicNumarasi = PageID - GosterilenSayfaSayisi;
            SayfaBitisNumarasi = PageID + GosterilenSayfaSayisi;
            //--------------------------------------------------------------------------------
            if (SayfaBaslangicNumarasi < GosterilenSayfaSayisi)
            {
                SayfaBaslangicNumarasi = 1;
                SayfaBitisNumarasi = (GosterilenSayfaSayisi * 2) + 1;
            }
            //--------------------------------------------------------------------------------
            if (SayfaBitisNumarasi > SayfaSayisi)
            {
                SayfaBitisNumarasi = SayfaSayisi;
                SayfaBaslangicNumarasi = SayfaBitisNumarasi - (GosterilenSayfaSayisi * 2);
            }
            //--------------------------------------------------------------------------------
            if (SayfaBaslangicNumarasi < 1)
            {
                SayfaBaslangicNumarasi = 1;
            }
            //--------------------------------------------------------------------------------
            if (PageID > 1)
            {
                shtml.Append("<a href=\"" + yol + "&PageID=1\" class=\"basaDon\" title=\"1\" style=\"background-image:url('/images/basa-don-icon.png')\"></a>");
                shtml.Append("<a href=\"" + yol + "&PageID=" + (PageID - 1) + "\" title=\"" + (PageID - 1) + "\"  class=\"basaDon\" style=\"background-image:url('/images/geri.png')\"></a>");
            }
            //--------------------------------------------------------------------------------
            for (int i = SayfaBaslangicNumarasi; i < SayfaBitisNumarasi + 1; i++)
            {
                if (PageID == i)
                {
                    shtml.Append("<a href=\"javascript:;\" class=\"sayiBg secili\" title=\"" + i.ToString() + "\">" + i.ToString() + "</a>");
                }
                else
                {
                    shtml.Append("<a href=\"" + yol + "&PageID=" + i + "\" class=\"sayiBg\" title=\"" + i.ToString() + "\">" + i.ToString() + "</a>");
                }
            }
            //--------------------------------------------------------------------------------
            if (PageID < SayfaSayisi)
            {
                if (PageID == 0)
                { shtml.Append("<a href=\"" + yol + "&PageID=" + (PageID + 2) + "\" class=\"basaDon\"  title=\"" + (PageID + 2) + "\" style=\"background-image:url('/images/ileri.png')\"></a>"); }
                else
                { shtml.Append("<a href=\"" + yol + "&PageID=" + (PageID + 1) + "\" class=\"basaDon\"  title=\"" + (PageID + 1) + "\" style=\"background-image:url('/images/ileri.png')\"></a>"); }
                //-----------------------------------------------------------------------------------------------------------------------------------------
                shtml.Append(" <a href=\"" + yol + "&PageID=" + SayfaSayisi + "\" class=\"basaDon\"  title=\"" + SayfaSayisi + "\"><img src=\"/images/son.png\" alt=\"" + SayfaSayisi + "\" /></a>");
                //-----------------------------------------------------------------------------------------------------------------------------------------
            }
            //--------------------------------------------------------------------------------
        }
        return shtml.ToString();
    }
    //--------------------------------------------------------------------------------------------

    public static String sayfalamayap(string sql, int SayfadakiKayitSayisi, int PageID, string yol)
    {
        int GosterilenSayfaSayisi = 2;
        int ToplamKayitSayisi = 0;
        int SayfaBaslangicNumarasi = 0;
        int SayfaBitisNumarasi = 0;
        int SayfaSayisi;
        //------------------------------------------------------------------
        DataTable rs = Baglanti.GetDataTable(sql);
        ToplamKayitSayisi = Convert.ToInt32(rs.Rows.Count);
        SayfaSayisi = Convert.ToInt32(Math.Ceiling(Convert.ToDouble(ToplamKayitSayisi) / Convert.ToDouble(SayfadakiKayitSayisi)));
        //------------------------------------------------------------------
        System.Text.StringBuilder shtml = new System.Text.StringBuilder();
        if (SayfaSayisi > 1)
        {

            SayfaBaslangicNumarasi = PageID - GosterilenSayfaSayisi;
            SayfaBitisNumarasi = PageID + GosterilenSayfaSayisi;
            //--------------------------------------------------------------------------------
            if (SayfaBaslangicNumarasi < GosterilenSayfaSayisi)
            {
                SayfaBaslangicNumarasi = 1;
                SayfaBitisNumarasi = (GosterilenSayfaSayisi * 2) + 1;
            }
            //--------------------------------------------------------------------------------
            if (SayfaBitisNumarasi > SayfaSayisi)
            {
                SayfaBitisNumarasi = SayfaSayisi;
                SayfaBaslangicNumarasi = SayfaBitisNumarasi - (GosterilenSayfaSayisi * 2);
            }
            //--------------------------------------------------------------------------------
            if (SayfaBaslangicNumarasi < 1)
            {
                SayfaBaslangicNumarasi = 1;
            }
            //--------------------------------------------------------------------------------
            if (PageID > 1)
            {
                shtml.Append("<a href=\"" + yol + "1.htm\" class=\"basaDon\" title=\"1\" style=\"background-image:url('/images/basa-don-icon.png')\"></a>");
                shtml.Append("<a href=\"" + yol + "" + (PageID - 1) + ".htm\" class=\"basaDon\" title=\"" + (PageID - 1) + "\" style=\"background-image:url('/images/geri.png')\"></a>");
            }
            //--------------------------------------------------------------------------------
            for (int i = SayfaBaslangicNumarasi; i < SayfaBitisNumarasi + 1; i++)
            {
                if (PageID == i)
                {
                    shtml.Append("<a href=\"javascript:;\" class=\"sayiBg secili\" title=\"" + i.ToString() + "\">" + i.ToString() + "</a>");
                }
                else
                {
                    shtml.Append("<a href=\"" + yol + "" + i + ".htm\" class=\"sayiBg\" title=\"" + i.ToString() + "\">" + i.ToString() + "</a>");
                }
            }
            //--------------------------------------------------------------------------------
            if (PageID < SayfaSayisi)
            {
                //--------------------------------------------------------------------------------------------------------------------
                if (PageID == 0)
                { shtml.Append("<a href=\"" + yol + "" + (PageID + 2) + ".htm\" class=\"basaDon\" title=\"" + (PageID + 2) + "\" style=\"background-image:url('/images/ileri.png')\"></a>"); }
                else
                { shtml.Append("<a href=\"" + yol + "" + (PageID + 1) + ".htm\" class=\"basaDon\" title=\"" + (PageID + 1) + "\" style=\"background-image:url('/images/ileri.png')\"></a>"); }
                //--------------------------------------------------------------------------------------------------------------------
                shtml.Append(" <a href=\"" + yol + "" + SayfaSayisi + ".htm\" class=\"basaDon\" title=\"" + SayfaSayisi + "\"><img src=\"/images/son.png\" alt=\"" + SayfaSayisi + "\" /></a>");
            }
            //--------------------------------------------------------------------------------
        }
        return shtml.ToString();
    }
    //--------------------------------------------------------------------------------------------
    public static string KategoriCokluBannerResimSec(string id)
    {
        string geridonenveri = null;
        //----------------------------------------------------------------------------------------
        string AtkifID = islem.CokluEnUstKategoriIDSec(Convert.ToInt32(id));
        string AtkifUID = islem.VeriGetirSec(id, "Genel_Kategori", "UstKategoriID", "id");
        //----------------------------------------------------------------------------------------
        if (!String.IsNullOrEmpty(islem.VeriGetirSec(id, "Genel_Kategori_Galeri", "Resim", "id")))
        { id = id; }
        else if (!String.IsNullOrEmpty(islem.VeriGetirSec(AtkifUID, "Genel_Kategori_Galeri", "Resim", "id")))
        { id = AtkifUID; }
        else if (!String.IsNullOrEmpty(islem.VeriGetirSec(AtkifID, "Genel_Kategori_Galeri", "Resim", "id")))
        { id = AtkifID; }
        //----------------------------------------------------------------------------------------

        DataRow rs = Baglanti.GetDataRow("select CatId,Resim from Tbl_Genel_Kategori_Galeri where CatId=" + Convert.ToInt32(id) + " and Resim<>'' and  Onay=1");
        if (rs != null)
        { geridonenveri = rs["CatId"].ToString(); }
        //---------------------------------------------------------------------------------
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------
    //--------------------------Sayfalama İşlemleri Son -------------------------------------------
    //----------------------------------------------------------------------------------------------



    //----------------------------------------------------------------------------------------------
    //--------------------------Croplama İşlemleri Başlangıç---------------------------------------
    //----------------------------------------------------------------------------------------------

    public static string CropGenelMenuKategoriSec(string ResimTuru, int CatID, int ResimDurum)
    {
        string geridonenveri = "0";
        //-------------------------------------------------------------------------------------
        string AlanAdi = "";
        if (ResimDurum == 0)
        {
            if (ResimTuru == "b_")
            { AlanAdi = "MenuBuyukResimGenislik"; }
            else if (ResimTuru == "o_")
            { AlanAdi = "MenuOrtaResimGenislik"; }
            else if (ResimTuru == "k_")
            { AlanAdi = "MenuKucukResimGenislik"; }
        }
        else if (ResimDurum == 1)
        {
            if (ResimTuru == "b_")
            { AlanAdi = "MenuBuyukResimYukseklik"; }
            else if (ResimTuru == "o_")
            { AlanAdi = "MenuOrtaResimYukseklik"; }
            else if (ResimTuru == "k_")
            { AlanAdi = "MenuKucukResimYukseklik"; }
        }
        //-------------------------------------------------------------------------------------
        DataRow rsk = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + ",UstKategoriID,SayfaTurID from Tbl_Genel_Kategori Where id=" + Convert.ToInt32(CatID) + "");
        if (rsk != null)
        {
            //---------------------------------------------------------------------------------
            if (Convert.ToInt32(rsk["" + AlanAdi + ""]) > 0)
            { geridonenveri = rsk["" + AlanAdi + ""].ToString(); }
            else if (Convert.ToInt32(rsk["" + AlanAdi + ""]) == 0)
            {
                DataRow rsenust = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + " from Tbl_Genel_Kategori Where id=" + Fonksiyon.SqlInjectionTemizle(islem.CssNoCokluKategoriIDSec(Convert.ToInt32(rsk["UstKategoriID"]), "0")) + "");
                if (rsenust != null)
                {
                    if (Convert.ToInt32(rsenust[AlanAdi]) > 0)
                    { geridonenveri = rsenust["" + AlanAdi + ""].ToString(); }
                    else
                    {
                        DataRow rsst = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + " from Tbl_Zm_SayfaTurleri Where id=" + Convert.ToInt32(rsk["SayfaTurID"]) + "");
                        if (rsst != null)
                        {
                            if (Convert.ToInt32(rsst[AlanAdi]) > 0)
                            { geridonenveri = rsst["" + AlanAdi + ""].ToString(); }
                        }
                    }
                }
                else
                {
                    DataRow rsst = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + " from Tbl_Zm_SayfaTurleri Where id=" + Convert.ToInt32(rsk["SayfaTurID"]) + "");
                    if (rsst != null)
                    {
                        if (Convert.ToInt32(rsst[AlanAdi]) > 0)
                        { geridonenveri = rsst["" + AlanAdi + ""].ToString(); }
                    }
                }
            }
            //---------------------------------------------------------------------------------
        }
        //-------------------------------------------------------------------------------
        if (geridonenveri == "0")
        {
            if (ResimTuru == "b_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.MenuBuyukResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.MenuBuyukResimYukseklik.ToString(); }
            }
            else if (ResimTuru == "o_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.MenuOrtaResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.MenuOrtaResimYukseklik.ToString(); }
            }
            else if (ResimTuru == "k_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.MenuKucukResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.MenuKucukResimYukseklik.ToString(); }
            }
        }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string CropGenelKayitlarSec(string ResimTuru, int CatID, int ResimDurum)
    {
        string geridonenveri = "0";
        //-------------------------------------------------------------------------------------
        string AlanAdi = "";
        if (ResimDurum == 0)
        {
            if (ResimTuru == "b_")
            { AlanAdi = "BuyukResimGenislik"; }
            else if (ResimTuru == "o_")
            { AlanAdi = "OrtaResimGenislik"; }
            else if (ResimTuru == "k_")
            { AlanAdi = "KucukResimGenislik"; }
        }
        else if (ResimDurum == 1)
        {
            if (ResimTuru == "b_")
            { AlanAdi = "BuyukResimYukseklik"; }
            else if (ResimTuru == "o_")
            { AlanAdi = "OrtaResimYukseklik"; }
            else if (ResimTuru == "k_")
            { AlanAdi = "KucukResimYukseklik"; }
        }
        //-------------------------------------------------------------------------------------
        DataRow rsk = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + ",UstKategoriID,SayfaTurID from Tbl_Genel_Kategori Where id=" + Convert.ToInt32(CatID) + "");
        if (rsk != null)
        {
            //---------------------------------------------------------------------------------
            if (Convert.ToInt32(rsk["" + AlanAdi + ""]) > 0)
            { geridonenveri = rsk["" + AlanAdi + ""].ToString(); }
            else if (Convert.ToInt32(rsk["" + AlanAdi + ""]) == 0)
            {
                DataRow rsenust = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + " from Tbl_Genel_Kategori Where id=" + Fonksiyon.SqlInjectionTemizle(islem.CssNoCokluKategoriIDSec(Convert.ToInt32(rsk["UstKategoriID"]), "0")) + "");
                if (rsenust != null)
                {
                    if (Convert.ToInt32(rsenust[AlanAdi]) > 0)
                    { geridonenveri = rsenust["" + AlanAdi + ""].ToString(); }
                    else
                    {
                        DataRow rsst = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + " from Tbl_Zm_SayfaTurleri Where id=" + Convert.ToInt32(rsk["SayfaTurID"]) + "");
                        if (rsst != null)
                        {
                            if (Convert.ToInt32(rsst[AlanAdi]) > 0)
                            { geridonenveri = rsst["" + AlanAdi + ""].ToString(); }
                        }
                    }
                }
                else
                {
                    DataRow rsst = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + " from Tbl_Zm_SayfaTurleri Where id=" + Convert.ToInt32(rsk["SayfaTurID"]) + "");
                    if (rsst != null)
                    {
                        if (Convert.ToInt32(rsst[AlanAdi]) > 0)
                        { geridonenveri = rsst["" + AlanAdi + ""].ToString(); }
                    }
                }
            }
        }
        //-------------------------------------------------------------------------------
        if (geridonenveri == "0")
        {
            if (ResimTuru == "b_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.BuyukResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.BuyukResimYukseklik.ToString(); }
            }
            else if (ResimTuru == "o_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.OrtaResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.OrtaResimYukseklik.ToString(); }
            }
            else if (ResimTuru == "k_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.KucukResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.KucukResimYukseklik.ToString(); }
            }
        }
        //---------------------------------------------------------------------------------
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string CropGenelCokluKayitlarSec(string ResimTuru, int CatID, int ResimDurum, string TabloAdi)
    {
        string geridonenveri = "0";
        //-------------------------------------------------------------------------------------
        CatID = Convert.ToInt32(islem.VeriGetirSec(CatID.ToString(), TabloAdi, "CatID", "id"));
        //-------------------------------------------------------------------------------------
        string AlanAdi = "";
        if (ResimDurum == 0)
        {
            if (ResimTuru == "b_")
            { AlanAdi = "CokluBuyukResimGenislik"; }
            else if (ResimTuru == "o_")
            { AlanAdi = "CokluOrtaResimGenislik"; }
            else if (ResimTuru == "k_")
            { AlanAdi = "CokluKucukResimGenislik"; }
        }
        else if (ResimDurum == 1)
        {
            if (ResimTuru == "b_")
            { AlanAdi = "CokluBuyukResimYukseklik"; }
            else if (ResimTuru == "o_")
            { AlanAdi = "CokluOrtaResimYukseklik"; }
            else if (ResimTuru == "k_")
            { AlanAdi = "CokluKucukResimYukseklik"; }
        }
        //-------------------------------------------------------------------------------------
        DataRow rsk = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + ",UstKategoriID,SayfaTurID from Tbl_Genel_Kategori Where id=" + Convert.ToInt32(CatID) + "");
        if (rsk != null)
        {
            //---------------------------------------------------------------------------------
            if (Convert.ToInt32(rsk["" + AlanAdi + ""]) > 0)
            { geridonenveri = rsk["" + AlanAdi + ""].ToString(); }
            else if (Convert.ToInt32(rsk["" + AlanAdi + ""]) == 0)
            {
                DataRow rsenust = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + " from Tbl_Genel_Kategori Where id=" + Fonksiyon.SqlInjectionTemizle(islem.CssNoCokluKategoriIDSec(Convert.ToInt32(rsk["UstKategoriID"]), "0")) + "");
                if (rsenust != null)
                {
                    if (Convert.ToInt32(rsenust[AlanAdi]) > 0)
                    { geridonenveri = rsenust["" + AlanAdi + ""].ToString(); }
                    else
                    {
                        DataRow rsst = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + " from Tbl_Zm_SayfaTurleri Where id=" + Convert.ToInt32(rsk["SayfaTurID"]) + "");
                        if (rsst != null)
                        {
                            if (Convert.ToInt32(rsst[AlanAdi]) > 0)
                            { geridonenveri = rsst["" + AlanAdi + ""].ToString(); }
                        }
                    }
                }
                else
                {
                    DataRow rsst = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + " from Tbl_Zm_SayfaTurleri Where id=" + Convert.ToInt32(rsk["SayfaTurID"]) + "");
                    if (rsst != null)
                    {
                        if (Convert.ToInt32(rsst[AlanAdi]) > 0)
                        { geridonenveri = rsst["" + AlanAdi + ""].ToString(); }
                    }
                }
            }
        }
        //-------------------------------------------------------------------------
        if (geridonenveri == "0")
        {
            if (ResimTuru == "b_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.CokluBuyukResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.CokluBuyukResimYukseklik.ToString(); }
            }
            else if (ResimTuru == "o_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.CokluOrtaResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.CokluOrtaResimYukseklik.ToString(); }
            }
            else if (ResimTuru == "k_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.CokluKucukResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.CokluKucukResimYukseklik.ToString(); }
            }
        }
        //---------------------------------------------------------------------------------
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string CropGenelBannerSec(string ResimTuru, int CatID, int ResimDurum)
    {
        string geridonenveri = "0";
        if (ResimTuru == "b_")
        {
            if (ResimDurum == 0)
            { geridonenveri = Ayarlar.BannerBuyukResimGenislik.ToString(); }
            else if (ResimDurum == 1)
            { geridonenveri = Ayarlar.BannerBuyukResimYukseklik.ToString(); }
        }
        else if (ResimTuru == "o_")
        {
            if (ResimDurum == 0)
            { geridonenveri = Ayarlar.BannerOrtaResimGenislik.ToString(); }
            else if (ResimDurum == 1)
            { geridonenveri = Ayarlar.BannerOrtaResimYukseklik.ToString(); }
        }
        else if (ResimTuru == "k_")
        {
            if (ResimDurum == 0)
            { geridonenveri = Ayarlar.BannerKucukResimGenislik.ToString(); }
            else if (ResimDurum == 1)
            { geridonenveri = Ayarlar.BannerKucukResimYukseklik.ToString(); }
        }
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string CropGenelSayfaTurKayitlarSec(string SayfaTurID, string ResimTuru, int id, int ResimDurum)
    {
        string geridonenveri = "0";
        //-------------------------------------------------------------------------------------
        string AlanAdi = "";
        if (ResimDurum == 0)
        {
            if (ResimTuru == "b_")
            { AlanAdi = "BuyukResimGenislik"; }
            else if (ResimTuru == "o_")
            { AlanAdi = "OrtaResimGenislik"; }
            else if (ResimTuru == "k_")
            { AlanAdi = "KucukResimGenislik"; }
        }
        else if (ResimDurum == 1)
        {
            if (ResimTuru == "b_")
            { AlanAdi = "BuyukResimYukseklik"; }
            else if (ResimTuru == "o_")
            { AlanAdi = "OrtaResimYukseklik"; }
            else if (ResimTuru == "k_")
            { AlanAdi = "KucukResimYukseklik"; }
        }
        //-------------------------------------------------------------------------------------

        //-------------------------------------------------------------------------------------
        DataRow rsst = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + " from Tbl_Zm_SayfaTurleri Where id=" + Convert.ToInt32(SayfaTurID) + "");
        if (rsst != null)
        {
            if (Convert.ToInt32(rsst[AlanAdi]) > 0)
            { geridonenveri = rsst["" + AlanAdi + ""].ToString(); }
        }
        //-------------------------------------------------------------------------------------

        //-------------------------------------------------------------------------------------
        if (geridonenveri == "0")
        {
            if (ResimTuru == "b_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.BuyukResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.BuyukResimYukseklik.ToString(); }
            }
            else if (ResimTuru == "o_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.OrtaResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.OrtaResimYukseklik.ToString(); }
            }
            else if (ResimTuru == "k_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.KucukResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.KucukResimYukseklik.ToString(); }
            }
        }
        //---------------------------------------------------------------------------------
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string CropGenelCokluSayfaTurKayitlarSec(string SayfaTurID, string ResimTuru, int ResimDurum)
    {
        string geridonenveri = "0";
        //-------------------------------------------------------------------------------------
        string AlanAdi = "";
        if (ResimDurum == 0)
        {
            if (ResimTuru == "b_")
            { AlanAdi = "CokluBuyukResimGenislik"; }
            else if (ResimTuru == "o_")
            { AlanAdi = "CokluOrtaResimGenislik"; }
            else if (ResimTuru == "k_")
            { AlanAdi = "CokluKucukResimGenislik"; }
        }
        else if (ResimDurum == 1)
        {
            if (ResimTuru == "b_")
            { AlanAdi = "CokluBuyukResimYukseklik"; }
            else if (ResimTuru == "o_")
            { AlanAdi = "CokluOrtaResimYukseklik"; }
            else if (ResimTuru == "k_")
            { AlanAdi = "CokluKucukResimYukseklik"; }
        }
        //-------------------------------------------------------------------------------------

        //-------------------------------------------------------------------------------------
        DataRow rsst = Baglanti.GetDataRow("Select " + Fonksiyon.SqlInjectionTemizle(AlanAdi) + " from Tbl_Zm_SayfaTurleri Where id=" + Convert.ToInt32(SayfaTurID) + "");
        if (rsst != null)
        {
            if (Convert.ToInt32(rsst[AlanAdi]) > 0)
            { geridonenveri = rsst["" + AlanAdi + ""].ToString(); }
        }
        //-------------------------------------------------------------------------------------

        //-------------------------------------------------------------------------
        if (geridonenveri == "0")
        {
            if (ResimTuru == "b_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.CokluBuyukResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.CokluBuyukResimYukseklik.ToString(); }
            }
            else if (ResimTuru == "o_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.CokluOrtaResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.CokluOrtaResimYukseklik.ToString(); }
            }
            else if (ResimTuru == "k_")
            {
                if (ResimDurum == 0)
                { geridonenveri = Ayarlar.CokluKucukResimGenislik.ToString(); }
                else if (ResimDurum == 1)
                { geridonenveri = Ayarlar.CokluKucukResimYukseklik.ToString(); }
            }
        }
        //---------------------------------------------------------------------------------
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static int CropBuyukGenislikResimKatiSec(int gelenveri)
    {
        int geridonenveri = 1;
        //---------------------------------------    
        try
        {
            //---------------------------------------    
            if (Ayarlar.CropBuyukGenislikResimKatiDurum == true)
            {
                DataRow rs = null;
                //---------------------------------------    
                rs = Baglanti.GetDataRow("Select CropBuyukGenislikResimKati from Tbl_Genel_Kategori where id='" + Convert.ToInt32(gelenveri) + "' and CropBuyukGenislikResimKati>1");
                //---------------------------------------    
                if (rs != null)
                { geridonenveri = Convert.ToInt32(rs[0].ToString()); }
                //---------------------------------------    
                else
                {
                    //--------------------------------------
                    string SayfaTurID = islem.VeriGetirSec(gelenveri.ToString(), "Genel_Kategori", "DevamiSayfaTurID", "id");
                    if ( SayfaTurID == "0" )
                    { SayfaTurID = islem.VeriGetirSec(gelenveri.ToString(), "Genel_Kategori", "SayfaTurID", "id"); }
                    //---------------------------------------    
                    rs = Baglanti.GetDataRow("Select CropBuyukGenislikResimKati from Tbl_Zm_SayfaTurleri where id='" + Convert.ToInt32(SayfaTurID) + "' and CropBuyukGenislikResimKati>1");
                    //---------------------------------------    
                    if (rs != null)
                    { geridonenveri = Convert.ToInt32(rs[0].ToString()); }
                    else
                    { geridonenveri = Ayarlar.CropBuyukGenislikResimKati; }
                    //---------------------------------------    
                }
                //---------------------------------------    
            }
            //---------------------------------------    
        }
        catch
        { geridonenveri = 1; }
        //---------------------------------------    
        return geridonenveri;
    }
    //----------------------------------------------------------------------------------------------


    //----------------------------------------------------------------------------------------------
    //--------------------------Croplama İşlemleri Son---------------------------------------------
    //----------------------------------------------------------------------------------------------




    //----------------------------------------------------------------------------------------------
    //--------------------------Siteye Özel Fonksiyonlar -------------------------------------------
    //----------------------------------------------------------------------------------------------




    //----------------------------------------------------------------------------------------------
    //--------------------------Siteye Özel Fonksiyonlar -------------------------------------------
    //----------------------------------------------------------------------------------------------

}