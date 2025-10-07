using System;
using System.Collections.Generic;
using System.Linq;
using System.Web;
public class Html
{
    //------------------------------------------------------------------------------------------------------------------
    public static string MailTo(string linkAd, string eposta, string cssClass)
    {
        string geridonenveri = null;
        geridonenveri = "<a href=\"mailto:" + eposta + "\" title=\"" + linkAd + "\" class=\"" + cssClass + "\">" + linkAd + "</a>";
        return geridonenveri;
    }
    //------------------------------------------------------------------------------------------------------------------
    
    public static string Div(object innerHtml)
    {
        return "<div>" + innerHtml + "</div>";
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string Div(object innerHtml, string cssClass)
    {
        return "<div class=\"" + cssClass + "\">" + innerHtml + "</div>";
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string Img(object imgSrc)
    {
        return "<img src=\"" + imgSrc + "\" alt=\"\" style=\"border:0px;\" />";
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string Img(object imgSrc, string cssClass)
    {
        return "<img src=\"" + imgSrc + "\" class=\"" + cssClass + "\" alt=\"\" style=\"border:0px;\" />";
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string Span(string innerHtml, string cssClass)
    {
        return "<span title=\"" + innerHtml + "\" class=\"" + cssClass + "\">" + innerHtml + "</span>";
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string Span(string innerHtml, string title, string cssClass)
    {
        return "<span title=\"" + title + "\" class=\"" + cssClass + "\">" + innerHtml + "</span>";
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string Span(string innerHtml, string cssClass, bool cssToConvertStyle)
    {
        return "<span title=\"" + innerHtml.Replace("</strong>", null).Replace("<strong>", null) + "\" " + cssClass + ">" + innerHtml + "</span>";
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string Bold(object text)
    {
        return "<strong>" + text + "</strong>";
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string H(object text, int h_numara)
    {
        return "<h" + h_numara + " title=\"" + text + "\">" + text + "</h" + h_numara + ">";
    }
    //------------------------------------------------------------------------------------------------------------------

    public static string Br(int br_Adet)
    {
        string geridonenveri = null;
        for (int i = 0; i < br_Adet; i++)
            geridonenveri += "<br/>";
        return geridonenveri;
    }
    //------------------------------------------------------------------------------------------------------------------

}