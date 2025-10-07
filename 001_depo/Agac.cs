using System;
using System.Collections.Generic;
using System.Linq;
using System.Web;
using System.Collections;
using System.Web.UI.WebControls;
using System.Web.UI;
public class Agac
{
    Literal ltrlAgac = new Literal();
    static ArrayList alAgac = new ArrayList();
    static string spPasif = null;
    //---------------------------------------------------------

    private static string AyracPanel
    { get { return "<li> / </li>"; } }
    //---------------------------------------------------------

    private static string AyracSite
    { get { return "<i class='pagenav-arrow'></i>"; } }
    //---------------------------------------------------------

    public static void Add(string linkAd, string linkUrl)
    { alAgac.Add(new ListItem(linkAd, linkUrl)); }
    //---------------------------------------------------------

    public static void Add(object linkAd, bool pasif)
    { spPasif = linkAd.ToString(); }
    //---------------------------------------------------------

    public static void RegisterAgacPanel(Literal ltrlAgac)
    {
        try
        {
            ltrlAgac.Text = Link(PanelAnasayfa, "l_agac");
            int _lnkCont = alAgac.Count;
            for (int i = 0; i < _lnkCont; i++)
            {
                ListItem li = (ListItem)alAgac[i];
                if (i == (_lnkCont - 1))
                {
                    ltrlAgac.Text += AyracPanel;
                    ltrlAgac.Text += LinkCurrent(li.Text);
                }
                else
                {
                    ltrlAgac.Text += AyracPanel;
                    ltrlAgac.Text += Link(li.Text, li.Value, "l_agac");
                }
            }
            alAgac.Clear();
        }
        catch { }
    }
    //---------------------------------------------------------

    public static void RegisterAgacSite(Literal ltrlAgac)
    {
        try
        {
            ltrlAgac.Text = LinkSite(PanelAnasayfaSite, "", PanelAnasayfaSite.ToString());
            int _lnkCont = alAgac.Count;
            for (int i = 0; i < _lnkCont; i++)
            {   
                ListItem li = (ListItem)alAgac[i];
                if (i == (_lnkCont - 1))
                {
                    ltrlAgac.Text += AyracSite;
                    //ltrlAgac.Text += LinkSite(li.Text, li.Value, islem.AgacHrefTitleSec(li.Value, li.Text));
                    ltrlAgac.Text += LinkCurrentSite(li.Text, islem.AgacHrefTitleSec(li.Value, li.Text));
                }
                else
                {
                    ltrlAgac.Text += AyracSite;
                    ltrlAgac.Text += LinkSite(li.Text, li.Value, "", islem.AgacHrefTitleSec(li.Value, li.Text));
                }
            }
            alAgac.Clear();
        }
        catch { }
    }
    //---------------------------------------------------------
    
    static string[] GelenVeri = new string[3];
    //---------------------------------------------------------

    public static string[] PanelAnasayfa
    {
        get
        {
            GelenVeri[0] = "Giriş";
            GelenVeri[1] = "Admin.aspx";
            GelenVeri[2] = "Admin.aspx";
            return GelenVeri;
        }
    }
    //---------------------------------------------------------

    public static string[] PanelAnasayfaSite
    {
        get
        {
            GelenVeri[0] = Ayarlar.FirmaAdi;
            GelenVeri[1] = "/";
            GelenVeri[2] = "/";
            return GelenVeri;
        }
    }
    //---------------------------------------------------------

    public static string ResolveUrl(string relativeUrl)
    {
        System.Web.UI.Page p = HttpContext.Current.Handler as System.Web.UI.Page;
        return p.ResolveUrl(relativeUrl);
    }
    //---------------------------------------------------------

    public static string Link(string linkAd, string url, string cssClass)
    {
        string geridonenveri = null;
        geridonenveri = "<li><a href=\"" + url + "\" title=\"" + linkAd + "\" class=\"" + cssClass + "\">" + linkAd + "</a></li>";
        return geridonenveri;
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string Link(string linkAd, string url)
    {
        string geridonenveri = null;
        geridonenveri = "<a href=\"" + url + "\" title=\"" + linkAd + "\">" + linkAd + "</a>";
        return geridonenveri;
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string Link(string[] sayfaAd, string cssClass)
    {
        string geridonenveri = null;
        geridonenveri = "<a href=\"" + sayfaAd[2] + "\" title=\"" + sayfaAd[0] + "\" class=\"" + cssClass + "\">" + sayfaAd[0] + "</a>";
        return geridonenveri;
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string LinkCurrent(string linkAd)
    {
        string geridonenveri = null;
        geridonenveri = "<li class=\"current\">" + linkAd + "</li/>";
        return geridonenveri;
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string LinkSite(string linkAd, string url, string cssClass, string linkAdTitle)
    {
        string geridonenveri = null;
        if (url == "javascript:;" || url == "#")
        { geridonenveri = "<a href=\"javascript:;\" title=\"" + linkAdTitle + "\" >" + linkAd + "</a>"; }
        else
        { geridonenveri = "<a href=\"" + url + "\" title=\"" + linkAdTitle + "\" >" + linkAd + "</a>"; }
        return geridonenveri;
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string LinkSite(string linkAd, string url, string linkAdTitle)
    {
        string geridonenveri = null;
        if (url == "javascript:;" || url == "#")
        { geridonenveri = "<a href=\"javascript:;\" title=\"" + linkAdTitle + "\">" + linkAd + "</a>"; }
        else
        { geridonenveri = "<a href=\"" + url + "\" title=\"" + linkAdTitle + "\">" + linkAd + "</a>"; }
        return geridonenveri;
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string LinkSite(string[] sayfaAd, string cssClass, string linkAdTitle)
    {
        string geridonenveri = null;
        if (sayfaAd[2] == "javascript:;" || sayfaAd[2] == "#")
        { geridonenveri = "<a href=\"javascript:;\" title=\"" + sayfaAd[0] + "\" >" + sayfaAd[0] + "</a>"; }
        else
        { geridonenveri = "<a href=\"" + sayfaAd[2] + "\" title=\"" + sayfaAd[0] + "\" >" + sayfaAd[0] + "</a>"; }
        return geridonenveri;
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string LinkCurrentSite(string linkAd, string linkAdTitle)
    {
        string geridonenveri = null;
        geridonenveri = "<a class=\"active\">" + linkAd + "</a>";
        return geridonenveri;
    }

    //------------------------------------------------------------------------------------------------------------------

}