using System;
using System.Data;
using System.Collections.Generic;
using System.Linq;
using System.Web;
using System.Web.UI.WebControls;
using System.Web.Mail;
using System.Text;
using System.Drawing;
using System.Drawing.Drawing2D;
using System.Drawing.Imaging;
using System.IO;
using System.Threading;
using System.Diagnostics;
using System.Web.UI.HtmlControls;
using System.Web.UI;
using System.Text.RegularExpressions;
using System.Drawing.Text;
using System.Net;
using Newtonsoft.Json;

public class Fonksiyon
{

    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------
    //--------------------------Koruma, Temizleme ve Yönlendirme------------------------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------

    public static string SqlInjectionTemizle(string gelenveri)
    {

        if (gelenveri != null)
        {
            string gvCopy = gelenveri.ToLowerInvariant();
            string[,] arr = new string[,] 
            { 
                { "'", "´" },{ "%27", "´" }, { "union", "" }, { "select", "" }, { "update", "" }, { "insert", "" }, { "delete", "" }, { "drop", "" }, { "+", "" }, { "into", "" },{ "where", "" }, { "order by", "" }, { "chr", "" }, { "isnull", "" }, { "xtype", "" }, { "sysobject", "" }, { "syscolumns", "" },{ "convert", "" }, { "db_name", "" }, { "@@", "" }, { "@var", "" }, { "declare", "" }, { "execute", "" }, { "having", "" }, { "1=1", "" }, { "exec", "" }, { "cmdshell", "" }, { "master", "" }, { "cmd", "" }, { "--", "" }, { "xp_", "" }, { ";", "" },  { "#", "" }, { "%", "" }, { "(", "" }, { ")", "" }, { "/*", "" }, { "*/", "" }, { "<", "" }, { ">", "" }, { "[", "" }, { "]", "" }, { "\"", "" }, { "½", "" }, { "^", "" }, { "{", "" }, { "}", "" }, { "~", "" }, { "|", "" }, { "*", "" }
            };
            int abc = -1;
            for (int i = 0; i < arr.Length / 2; i++)
            {
                abc = gvCopy.IndexOf(arr[i, 0]);
                if (abc > -1)
                {
                bastan:
                    gelenveri = gelenveri.Substring(0, abc) + arr[i, 1] + gelenveri.Substring(abc + arr[i, 0].Length, gelenveri.Length - abc - arr[i, 0].Length);
                    gvCopy = gvCopy.Substring(0, abc) + arr[i, 1] + gvCopy.Substring(abc + arr[i, 0].Length, gvCopy.Length - abc - arr[i, 0].Length);
                    abc = gvCopy.IndexOf(arr[i, 0]);
                    if (abc > -1) goto bastan;
                }
            }
        }
        return gelenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string KisaSqlInjectionTemizle(string gelenveri)
    {

        if (gelenveri != null)
        {
            string gvCopy = gelenveri.ToLowerInvariant();
            string[,] arr = new string[,] 
            { 
                { "'", "´" },{ "%27", "´" }, { "union", "" }, { "select", "" }, { "update", "" }, { "insert", "" }, { "delete", "" }, { "order by", "" }, { "chr", "" }, { "--", "" }, { ";", "" }, { "%", "" }, { "/*", "" }, { "*/", "" }, { "<", "" }, { ">", "" }
            };
            int abc = -1;
            for (int i = 0; i < arr.Length / 2; i++)
            {
                abc = gvCopy.IndexOf(arr[i, 0]);
                if (abc > -1)
                {
                bastan:
                    gelenveri = gelenveri.Substring(0, abc) + arr[i, 1] + gelenveri.Substring(abc + arr[i, 0].Length, gelenveri.Length - abc - arr[i, 0].Length);
                    gvCopy = gvCopy.Substring(0, abc) + arr[i, 1] + gvCopy.Substring(abc + arr[i, 0].Length, gvCopy.Length - abc - arr[i, 0].Length);
                    abc = gvCopy.IndexOf(arr[i, 0]);
                    if (abc > -1) goto bastan;
                }
            }
        }
        return gelenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string RobotsTxtSqlInjectionTemizle(string gelenveri)
    {

        if (gelenveri != null)
        {
            string gvCopy = gelenveri.ToLowerInvariant();
            string[,] arr = new string[,] 
            { 
                { "'", "´" },{ "%27", "´" }, { "union", "" }, { "select", "" }, { "update", "" }, { "insert", "" }, { "delete", "" }, { "order by", "" }, { "chr", "" }, { "--", "" }, { "<", "" }, { ">", "" }
            };
            int abc = -1;
            for (int i = 0; i < arr.Length / 2; i++)
            {
                abc = gvCopy.IndexOf(arr[i, 0]);
                if (abc > -1)
                {
                bastan:
                    gelenveri = gelenveri.Substring(0, abc) + arr[i, 1] + gelenveri.Substring(abc + arr[i, 0].Length, gelenveri.Length - abc - arr[i, 0].Length);
                    gvCopy = gvCopy.Substring(0, abc) + arr[i, 1] + gvCopy.Substring(abc + arr[i, 0].Length, gvCopy.Length - abc - arr[i, 0].Length);
                    abc = gvCopy.IndexOf(arr[i, 0]);
                    if (abc > -1) goto bastan;
                }
            }
        }
        return gelenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string SqlTemizle(string gelenveri)
    {
        if (gelenveri != null)
        {
            string gvCopy = gelenveri.ToLowerInvariant();
            string[,] arr = new string[,] 
            { 
                { "'", "´" },{ "%27", "" }, { "union", "" }, { "--", "" }, { "/*", "" }, { "*/", "" }, { "xtype", "" }, { "sysobject", "" },{ "syscolumns", "" }, { "db_name", "" }, { "@@", "" }, { "@var", "" },{ "chr", "" }, { "1=1", "" }, { "cmdshell", "" }, { "cmd", "" }, { "xp_", "" }, { "order by", "" }, { "order+by", "" }, { "group by", "" }, { "group+by", "" },
                ////----------Duruma Göre İptal Edebilir--------
                { "select", "" }, { "delete", "" }, { "update", "" }, { "insert into", "" }, { "/add", "" }
                ////------------------------------------------
            };
            int abc = -1;
            for (int i = 0; i < arr.Length / 2; i++)
            {
                abc = gvCopy.IndexOf(arr[i, 0]);
                if (abc > -1)
                {
                bastan:
                    gelenveri = gelenveri.Substring(0, abc) + arr[i, 1] + gelenveri.Substring(abc + arr[i, 0].Length, gelenveri.Length - abc - arr[i, 0].Length);
                    gvCopy = gvCopy.Substring(0, abc) + arr[i, 1] + gvCopy.Substring(abc + arr[i, 0].Length, gvCopy.Length - abc - arr[i, 0].Length);
                    abc = gvCopy.IndexOf(arr[i, 0]);
                    if (abc > -1) goto bastan;
                }
            }
        }
        return gelenveri;

    }
    //----------------------------------------------------------------------------------------------

    public static string SqlReklamTemizle(string gelenveri)
    {
        if (gelenveri != null)
        {
            string gvCopy = gelenveri.ToLowerInvariant();
            string[,] arr = new string[,] 
            { 
                { "'", "´" },{ "%27", "" }, { "union", "" }, { "/*", "" }, { "*/", "" }, { "xtype", "" }, { "sysobject", "" },{ "syscolumns", "" }, { "db_name", "" }, { "@@", "" }, { "@var", "" },{ "chr", "" }, { "1=1", "" }, { "cmdshell", "" }, { "cmd", "" }, { "xp_", "" }, { "order by", "" }, { "order+by", "" }, { "group by", "" }, { "group+by", "" },
                ////----------Duruma Göre İptal Edebilir--------
                { "select", "" }, { "delete", "" }, { "update", "" }, { "insert into", "" }, { "/add", "" }
                ////------------------------------------------
            };
            int abc = -1;
            for (int i = 0; i < arr.Length / 2; i++)
            {
                abc = gvCopy.IndexOf(arr[i, 0]);
                if (abc > -1)
                {
                bastan:
                    gelenveri = gelenveri.Substring(0, abc) + arr[i, 1] + gelenveri.Substring(abc + arr[i, 0].Length, gelenveri.Length - abc - arr[i, 0].Length);
                    gvCopy = gvCopy.Substring(0, abc) + arr[i, 1] + gvCopy.Substring(abc + arr[i, 0].Length, gvCopy.Length - abc - arr[i, 0].Length);
                    abc = gvCopy.IndexOf(arr[i, 0]);
                    if (abc > -1) goto bastan;
                }
            }
        }
        return gelenveri;

    }
    //----------------------------------------------------------------------------------------------

    public static string SeoReplace(string gelenveri)
    {
        if (gelenveri != null)
        {
            gelenveri = gelenveri.Trim();
            string gvCopy = gelenveri.ToLowerInvariant().Trim();
            string[,] arr = new string[,] 
            { 
                { ",", "_" },{ "'", "_" },{ ":", "" },{ "%27", "" },{ "?", "" },{ "*", "" },{ "&#199;", "o" },{ "&#246;", "o" },{ "&#214;", "o" },{ "&#252;", "u" },{ "&#220;", "u" },{ "&#231;", "c" },{ "&#174;", "®" },{ "&amp;", "-" },{ "&nbsp;", "-" },{ " ", "-" },{ ";", "-" },{ "%20", "-" },{ "/", "-" },{ ".", "" },{ "ç", "c" },{ "Ç", "c" },{ "ğ", "g" },{ "Ğ", "g" },{ "İ", "i" },{ "I", "i" },{ "ı", "i" },{ "ö", "o" },{ "Ö", "o" },{ "ş", "s" },{ "Ş", "s" },{ "ü", "u" },{ "Ü", "u" },{ ".", "" },{ "’", "" },{ "'", "" },{ "(", "_" },{ ")", "_" },{ "/", "_" },{ "<", "_" },{ ">", "_" },{ "\"", "_" },{ "\\", "_" },{ "{", "_" },{ "}", "_" },{ "%", "_" },{ "&", "_" },{ "+", "_" },{ "//", "_" },{ "__", "_" },{ "³", "_" },{ "²", "2" },{ "“", null },{ "”", null },{ "’", null },{ "”", null },{ "&", "-" },{ "[^\\w]", "-" },{ "----", "-" },{ "---", "-" },{ "--", "-" },{ "[", "-" },{ "]", "-" },{ "½", "-" },{ "^", "-" },{ "~", "-" },{ "|", "-" },{ "*", "-" },{ "#", "-" },{ "%", "-" },{ "union", "" },{ "select", "" },{ "update", "" },{ "insert", "" },{ "delete", "" },{ "drop", "" },{ "into", "" },{ "where", "" },{ "order", "" },{ "chr", "" },{ "isnull", "" },{ "xtype", "" },{ "sysobject", "" },{ "syscolumns", "" },{ "convert", "" },{ "db_name", "" },{ "@@", "-" },{ "@var", "-" },{ "declare", "" },{ "execute", "" },{ "having", "" },{ "1=1", "-" },{ "exec", "" },{ "cmdshell", "" },{ "master", "" },{ "cmd", "-" },{ "xp_", "-" },
            };
            int abc = -1;
            for (int i = 0; i < arr.Length / 2; i++)
            {
                abc = gvCopy.IndexOf(arr[i, 0]);
                if (abc > -1)
                {
                bastan:
                    gelenveri = gelenveri.Substring(0, abc) + arr[i, 1] + gelenveri.Substring(abc + arr[i, 0].Length, gelenveri.Length - abc - arr[i, 0].Length);
                    gvCopy = gvCopy.Substring(0, abc) + arr[i, 1] + gvCopy.Substring(abc + arr[i, 0].Length, gvCopy.Length - abc - arr[i, 0].Length);
                    abc = gvCopy.IndexOf(arr[i, 0]);
                    if (abc > -1) goto bastan;
                }
            }
        }
        return gelenveri.ToLowerInvariant().Trim();
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimSeoReplace(string gelenveri)
    {
        if (gelenveri != null)
        {
            gelenveri = gelenveri.Trim();
            string gvCopy = gelenveri.ToLowerInvariant().Trim();
            string[,] arr = new string[,] 
            { 
                { ".", "-" },{ "_", "-" },{ ",", "-" },{ "'", "-" },{ ":", "" },{ "%27", "" },{ "?", "" },{ "*", "" },{ "&#199;", "o" },{ "&#246;", "o" },{ "&#214;", "o" },{ "&#252;", "u" },{ "&#220;", "u" },{ "&#231;", "c" },{ "&#174;", "®" },{ "&amp;", "-" },{ "&nbsp;", "-" },{ " ", "-" },{ ";", "-" },{ "%20", "-" },{ "/", "-" },{ ".", "" },{ "ç", "c" },{ "Ç", "c" },{ "ğ", "g" },{ "Ğ", "g" },{ "İ", "i" },{ "I", "i" },{ "ı", "i" },{ "ö", "o" },{ "Ö", "o" },{ "ş", "s" },{ "Ş", "s" },{ "ü", "u" },{ "Ü", "u" },{ ".", "" },{ "’", "" },{ "'", "" },{ "(", "-" },{ ")", "-" },{ "/", "-" },{ "<", "-" },{ ">", "-" },{ "\"", "-" },{ "\\", "-" },{ "{", "-" },{ "}", "-" },{ "%", "-" },{ "&", "-" },{ "+", "-" },{ "//", "-" },{ "--", "-" },{ "³", "-" },{ "²", "2" },{ "“", null },{ "”", null },{ "’", null },{ "”", null },{ "&", "-" },{ "[^\\w]", "-" },{ "----", "-" },{ "---", "-" },{ "--", "-" },{ "[", "-" },{ "]", "-" },{ "½", "-" },{ "^", "-" },{ "~", "-" },{ "|", "-" },{ "*", "-" },{ "#", "-" },{ "%", "-" },{ "union", "" },{ "select", "" },{ "update", "" },{ "insert", "" },{ "delete", "" },{ "drop", "" },{ "into", "" },{ "where", "" },{ "order", "" },{ "chr", "" },{ "isnull", "" },{ "xtype", "" },{ "sysobject", "" },{ "syscolumns", "" },{ "convert", "" },{ "db_name", "" },{ "@@", "-" },{ "@var", "-" },{ "declare", "" },{ "execute", "" },{ "having", "" },{ "1=1", "-" },{ "exec", "" },{ "cmdshell", "" },{ "master", "" },{ "cmd", "-" },{ "xp_", "-" },{ "--", "-" }
            };
            int abc = -1;
            for (int i = 0; i < arr.Length / 2; i++)
            {
                abc = gvCopy.IndexOf(arr[i, 0]);
                if (abc > -1)
                {
                bastan:
                    gelenveri = gelenveri.Substring(0, abc) + arr[i, 1] + gelenveri.Substring(abc + arr[i, 0].Length, gelenveri.Length - abc - arr[i, 0].Length);
                    gvCopy = gvCopy.Substring(0, abc) + arr[i, 1] + gvCopy.Substring(abc + arr[i, 0].Length, gvCopy.Length - abc - arr[i, 0].Length);
                    abc = gvCopy.IndexOf(arr[i, 0]);
                    if (abc > -1) goto bastan;
                }
            }
        }
        return gelenveri.ToLowerInvariant().Trim();
    }
    //----------------------------------------------------------------------------------------------

    public static string KisaveOzTemizle(string gelenveri)
    {
        if (gelenveri != null)
        {
            string gvCopy = gelenveri.ToLowerInvariant();
            string[,] arr = new string[,] 
            { 
                { "'", "" },{ "%27", "" }, { ";", "" }
            };
            int abc = -1;
            for (int i = 0; i < arr.Length / 2; i++)
            {
                abc = gvCopy.IndexOf(arr[i, 0]);
                if (abc > -1)
                {
                bastan:
                    gelenveri = gelenveri.Substring(0, abc) + arr[i, 1] + gelenveri.Substring(abc + arr[i, 0].Length, gelenveri.Length - abc - arr[i, 0].Length);
                    gvCopy = gvCopy.Substring(0, abc) + arr[i, 1] + gvCopy.Substring(abc + arr[i, 0].Length, gvCopy.Length - abc - arr[i, 0].Length);
                    abc = gvCopy.IndexOf(arr[i, 0]);
                    if (abc > -1) goto bastan;
                }
            }
        }
        return gelenveri;
    }
    //--------------------------------------------------------------------------------------------

    public static string XSSTemizle(string gelenveri)
    {
        if (gelenveri != null)
        {
            string gvCopy = gelenveri.ToLowerInvariant();
            string[,] arr = new string[,] 
            { 
                { "'", "" },{ "%27", "" },{ "<", "" },{ ">", "" },{ "&lt;", "" },{ "&gt;", "" },  { "script", "" }, { "cookie", "" }, { "document", "" }, { "src", "" }, { "alert", "" }, { "embed", "" }, { "object", "" }, { "applet", "" }, { "geturl", "" }, { "applet", "" }, { ";", "" }
            };
            int abc = -1;
            for (int i = 0; i < arr.Length / 2; i++)
            {
                abc = gvCopy.IndexOf(arr[i, 0]);
                if (abc > -1)
                {
                bastan:
                    gelenveri = gelenveri.Substring(0, abc) + arr[i, 1] + gelenveri.Substring(abc + arr[i, 0].Length, gelenveri.Length - abc - arr[i, 0].Length);
                    gvCopy = gvCopy.Substring(0, abc) + arr[i, 1] + gvCopy.Substring(abc + arr[i, 0].Length, gvCopy.Length - abc - arr[i, 0].Length);
                    abc = gvCopy.IndexOf(arr[i, 0]);
                    if (abc > -1) goto bastan;
                }
            }
        }
        return gelenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string BadRequestKontrol()
    {
        if (BadRequestizSiteKontrol() == false)
        {
            if (HttpContext.Current.Request.QueryString.Count > 0)
            { HttpContext.Current.Response.Redirect("/404.html"); }
        }
        return null;
    }
    //----------------------------------------------------------------------------------------------

    public static bool BadRequestizSiteKontrol()
    {
        bool donengelenveri = false;
        if (HttpContext.Current.Request.QueryString.ToString().IndexOf("facebook.com") != -1)
        { donengelenveri = true; }
        else if (HttpContext.Current.Request.QueryString.ToString().IndexOf("twiter.com") != -1)
        { donengelenveri = true; }
        return donengelenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string JqueryArama(string gelenveri)
    {
        if (gelenveri != null)
        {
            gelenveri = gelenveri.Replace("ã¶", "ö");
            gelenveri = gelenveri.Replace("Ã¶", "ö");
            gelenveri = gelenveri.Replace("Ä±", "ı");
            gelenveri = gelenveri.Replace("Ä°", "İ");
            gelenveri = gelenveri.Replace("Ã¼", "ü");
            gelenveri = gelenveri.Replace("Ã§", "ç");
            gelenveri = gelenveri.Replace("Ã", "Ö");
            gelenveri = gelenveri.Replace("Ã", "Ç");
            gelenveri = gelenveri.Replace("Ä", "ğ");
            gelenveri = gelenveri.Replace("Ä", "Ğ");
            gelenveri = gelenveri.Replace("å", "ş");
            gelenveri = gelenveri.Replace("Å", "Ş");
            gelenveri = gelenveri.Replace("Â", "");
            gelenveri = gelenveri.Replace("Ã", "Ü");
        }
        return gelenveri;
    }
    //----------------------------------------------------------------------------------------------

    public static string EditorKarakterDuzenleyici(string metin)
    {
        string Duzenlenmis = metin;
        if (metin != null)
        {
            Duzenlenmis = Duzenlenmis.Replace("&#304;", "I");
            Duzenlenmis = Duzenlenmis.Replace("&#305;", "i");
            Duzenlenmis = Duzenlenmis.Replace("&#214;", "Ö");
            Duzenlenmis = Duzenlenmis.Replace("&#246;", "ö");
            Duzenlenmis = Duzenlenmis.Replace("&Ouml;", "Ö");
            Duzenlenmis = Duzenlenmis.Replace("&ouml;", "ö");
            Duzenlenmis = Duzenlenmis.Replace("&#220;", "Ü");
            Duzenlenmis = Duzenlenmis.Replace("&#252;", "ü");
            Duzenlenmis = Duzenlenmis.Replace("&Uuml;", "Ü");
            Duzenlenmis = Duzenlenmis.Replace("&uuml;", "ü");
            Duzenlenmis = Duzenlenmis.Replace("&#199;", "Ç");
            Duzenlenmis = Duzenlenmis.Replace("&#231;", "ç");
            Duzenlenmis = Duzenlenmis.Replace("&Ccedil;", "Ç");
            Duzenlenmis = Duzenlenmis.Replace("&ccedil;", "ç");
            Duzenlenmis = Duzenlenmis.Replace("&#286;", "G");
            Duzenlenmis = Duzenlenmis.Replace("&#287;", "g");
            Duzenlenmis = Duzenlenmis.Replace("&#350;", "S");
            Duzenlenmis = Duzenlenmis.Replace("&#351;", "s");
            //------------------------------------------------
            Duzenlenmis = Duzenlenmis.Replace("&ldquo;", "“");
            Duzenlenmis = Duzenlenmis.Replace("&rdquo;", "”");

        }
        return Duzenlenmis;
    }
    //------------------------------------------------------------------------------------------------

    public static int UploadSqlTemizle(string gelenveri)
    {
        int veridurum = 0;
        //----------------------------------------------------------
        if (gelenveri != null)
        {
            gelenveri = gelenveri.ToLowerInvariant();
            //----------------------------------------------------------
            if (gelenveri.IndexOf("asp") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("aspx") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("php") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("css") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("js") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("html") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("cgi") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("xml") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("jsp") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("exe") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("cer") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf(";") != -1)
            { veridurum = 1; }
            //----------------------------------------------------------
            if (Ayarlar.ResimCiftNoktaDurum == true)
            {
                if (Convert.ToInt32(gelenveri.ToString().Split('.').Length) > 2)
                { veridurum = 1; }
            }
            //----------------------------------------------------------
        }
        //----------------------------------------------------------
        return veridurum;
        //----------------------------------------------------------
    }
    //----------------------------------------------------------------------------------------------

    public static int LogSqlInjectionTemizle(string gelenveri)
    {
        int veridurum = 0;
        if (gelenveri != null)
        {
            gelenveri = gelenveri.ToLowerInvariant();
            if (gelenveri.IndexOf("'") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("--;") != -1)
            { veridurum = 2; }
            else if (gelenveri.IndexOf("union") != -1)
            { veridurum = 3; }
            else if (gelenveri.IndexOf("@var") != -1)
            { veridurum = 4; }
            else if (gelenveri.IndexOf("@@") != -1)
            { veridurum = 5; }
            else if (gelenveri.IndexOf("/*") != -1)
            { veridurum = 6; }
            else if (gelenveri.IndexOf("*/") != -1)
            { veridurum = 7; }
            else if (gelenveri.IndexOf(";") != -1)
            { veridurum = 8; }
            else if (gelenveri.IndexOf("(") != -1)
            { veridurum = 9; }
            else if (gelenveri.IndexOf(")") != -1)
            { veridurum = 10; }
            else if (gelenveri.IndexOf("<") != -1)
            { veridurum = 11; }
            else if (gelenveri.IndexOf(">") != -1)
            { veridurum = 12; }
            else if (gelenveri.IndexOf("[") != -1)
            { veridurum = 13; }
            else if (gelenveri.IndexOf("]") != -1)
            { veridurum = 14; }
            else if (gelenveri.IndexOf("[") != -1)
            { veridurum = 15; }
            else if (gelenveri.IndexOf("{") != -1)
            { veridurum = 16; }
            else if (gelenveri.IndexOf("}") != -1)
            { veridurum = 17; }
            else if (gelenveri.IndexOf("~") != -1)
            { veridurum = 18; }
            else if (gelenveri.IndexOf("%27") != -1)
            { veridurum = 19; }
        }
        return veridurum;

    }
    //----------------------------------------------------------------------------------------------

    public static string ResimUzantisiBul(string gelenveri)
    {
        string veridurum = "";
        //----------------------------------------------------------
        if (gelenveri != null)
        {
            gelenveri = gelenveri.ToLowerInvariant();
            //------------------------------------------------------
            if (gelenveri.IndexOf(".jpg") != -1)
            { veridurum = ".jpg"; }
            else if (gelenveri.IndexOf(".jpeg") != -1)
            { veridurum = ".jpeg"; }
            else if (gelenveri.IndexOf(".gif") != -1)
            { veridurum = ".gif"; }
            else if (gelenveri.IndexOf(".png") != -1)
            { veridurum = ".png"; }
            //-------------------------------------------------------
        }
        //-----------------------------------------------------------
        return veridurum;
        //-----------------------------------------------------------
    }
    //----------------------------------------------------------------------------------------------

    private static Regex invalidXMLChars = new Regex(@"(?<![\uD800-\uDBFF])[\uDC00-\uDFFF]|[\uD800-\uDBFF](?![\uDC00-\uDFFF])|[\x00-\x08\x0B\x0C\x0E-\x1F\x7F-\x9F\uFEFF\uFFFE\uFFFF]", RegexOptions.Compiled);
    //----------------------------------------------------------------------------------------------

    public static string XmlTemizleme(string text)
    {
        if (!String.IsNullOrEmpty(text))
        { return invalidXMLChars.Replace(text, ""); }
        else
        { return null; }
    }
    //----------------------------------------------------------------------------------------------

    private static readonly Regex MobileRegex = new Regex(@"(nokia|sonyericsson|blackberry|IPHONE|samsung|sec-|windows ce|motorola|mot-|up.b|midp-)", RegexOptions.IgnoreCase | RegexOptions.Compiled);
    //----------------------------------------------------------------------------------------------

    public bool IsMobile
    {
        get
        {
            HttpRequest r = HttpContext.Current.Request;
            //--------------------------------------------
            if (r.Browser.IsMobileDevice)
            { return true; }
            //--------------------------------------------
            if (!string.IsNullOrEmpty(r.UserAgent) && MobileRegex.IsMatch(r.UserAgent))
            { return true; }
            //--------------------------------------------
            return false;
            //--------------------------------------------
        }
    }
    //----------------------------------------------------------------------------------------------

    public static int Mobile(string gelenveri)
    {
        int veridurum = 0;
        if (gelenveri != null)
        {
            gelenveri = gelenveri.ToLowerInvariant();
            if (gelenveri.IndexOf("text/vnd.wap.wml") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("android") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("ipad") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("symbian") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("series") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("nokia") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("mot-") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("motorola") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("lg-") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("lge") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("nec-") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("lg/") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("samsung") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("sie-") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("sec-") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("sgh-") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("sonyericsson") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("sharp") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("windows ce") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("portalmmm") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("o2-") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("docomo") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("philips") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("panasonic") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("sagem") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("smartphone") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("up.browser") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("up.link") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("slurp") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("spring") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("alcatel") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("sendo") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("blackberry") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("opera mini") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("opera 2") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("netfront") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("mobilephone mm") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("vodafone") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("avantgo") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("palmsource") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("siemens") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("toshiba") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("i-mobile") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("asus") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("ice") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("kwc") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("htc") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("softbank") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("mozilla/5.0 (x11; u; linux i686; en-us; rv:1.8.0.7) gecko/20060909 firefox/1.5.0.7") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("playstation") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("nitro") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("iphone") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("ipod") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("t-mobile") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("obigo") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("brew") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("yahooseeker") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("msmobot") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("novarra") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("skp") != -1)
            { veridurum = 1; }
            else if (gelenveri.IndexOf("openweb") != -1)
            { veridurum = 1; }
        }
        return veridurum;
    }
    //----------------------------------------------------------------------------------------------

    public bool MobilYonlendirme(string YonlendirmeAdresi)
    {
        //----------------------------------------------------------------
        bool durum = false;
        //----------------------------------------------------------------
        try
        {
            //--------------------------------------------------------------------------------
            if (HttpContext.Current.Request["orjinal"] == null && Convert.ToInt32(HttpContext.Current.Session["orjinalID"]) == 0)
            {
                int mobil = Fonksiyon.Mobile(HttpContext.Current.Request.ServerVariables["HTTP_USER_AGENT"]);
                if (mobil == 1)
                { HttpContext.Current.Response.Redirect(YonlendirmeAdresi); }

            }
            else { HttpContext.Current.Session["orjinalID"] = 1; }
            //-------------------------------------------------------------------------------
            durum = true;
        }
        catch {  }
        //----------------------------------------------------------------
        return durum;
        //----------------------------------------------------------------
    }
    //----------------------------------------------------------------------------------------------

    public static string SiteTryKontrol(string gelenveri)
    {
        try
        {
            if (LogSqlInjectionTemizle(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) != 0)
            {
                //-----------------------------------------------//
                PanelAyarlar.LogGenelRaporlari(0, "", "", 6);
                Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Sql İnjection Saldırısı", PanelAyarlar.LogSqlInjectionMailGonderme("Sql"));
                //-----------------------------------------------//
            }
            HttpContext.Current.Response.Redirect(ResolveUrl("~") + "404.html");
            return gelenveri;
        }
        catch { return null; }
    }
    //----------------------------------------------------------------------------------------------


    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------
    //------------------ Database .Net Componentle Baglantıları ve Listelemeler --------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------

    public static Repeater Repeater(Repeater Repeater, DataTable DataTable)
    { Repeater.DataSource = DataTable; Repeater.DataBind(); return Repeater; }
    //----------------------------------------------------------------------------------------------

    public static DataList DataList(DataList DataList, DataTable DataTable)
    { DataList.DataSource = DataTable; DataList.DataBind(); return DataList; }
    //----------------------------------------------------------------------------------------------

    public static void RadioButtonList(RadioButtonList RadioButtonList, DataTable DataTable)
    {
        try
        {
            RadioButtonList.Items.Clear();
            RadioButtonList.SelectedIndex = 0;
            RadioButtonList.DataSource = DataTable;
            RadioButtonList.DataTextField = DataTable.Columns[0].ToString();
            RadioButtonList.DataValueField = DataTable.Columns[0].ToString();
            RadioButtonList.DataBind();
        }
        catch { }
    }
    //----------------------------------------------------------------------------------------------

    public static void RadioButtonListBos(RadioButtonList RadioButtonList, DataTable DataTable)
    {
        try
        {
            RadioButtonList.Items.Clear();
            RadioButtonList.DataSource = DataTable;
            RadioButtonList.DataTextField = DataTable.Columns[0].ToString();
            RadioButtonList.DataValueField = DataTable.Columns[0].ToString();
            RadioButtonList.DataBind();
        }
        catch { }
    }
    //--------------------------------------------------------------------------------------------

    public static void ListBoxList(ListBox ListBoxList, DataTable DataTable)
    {
        try
        {
            ListBoxList.Items.Clear();
            ListBoxList.DataSource = DataTable;
            ListBoxList.DataTextField = DataTable.Columns[1].ToString();
            ListBoxList.DataValueField = DataTable.Columns[0].ToString();
            ListBoxList.DataBind();
        }
        catch { }
    }
    //----------------------------------------------------------------------------------------------

    public static void ListBoxListBos(ListBox ListBoxList, DataTable DataTable)
    {
        try
        {
            ListBoxList.Items.Clear();
            ListBoxList.DataSource = DataTable;
            ListBoxList.DataTextField = DataTable.Columns[1].ToString();
            ListBoxList.DataValueField = DataTable.Columns[0].ToString();
            ListBoxList.DataBind();
        }
        catch { }
    }
    //-------------------------------------------------------------------------------------------

    public static void DropDownList(DropDownList DropDownList, String ilkbaslik, DataTable DataTable)
    {
        try
        {
            DropDownList.Items.Clear();
            ListItem ilkveri = new ListItem(ilkbaslik.ToString(), "0", true); ilkveri.Selected = true;
            DropDownList.Items.Add(ilkveri);
            DropDownList.DataSource = DataTable;
            DropDownList.DataTextField = DataTable.Columns[1].ToString();
            DropDownList.DataValueField = DataTable.Columns[0].ToString();
            DropDownList.DataBind();
        }
        catch { }
    }
    //----------------------------------------------------------------------------------------------

    public static void DropDownListAyni(DropDownList DropDownList, String ilkbaslik, DataTable DataTable)
    {
        try
        {
            DropDownList.Items.Clear();
            ListItem ilkveri = new ListItem(ilkbaslik.ToString(), "0", true); ilkveri.Selected = true;
            DropDownList.Items.Add(ilkveri);
            DropDownList.DataSource = DataTable;
            DropDownList.DataTextField = DataTable.Columns[1].ToString();
            DropDownList.DataValueField = DataTable.Columns[1].ToString();
            DropDownList.DataBind();
        }
        catch { }
    }
    //----------------------------------------------------------------------------------------------

    public static void DropDownListBos(DropDownList DropDownList, String ilkbaslik, DataTable DataTable)
    {
        try
        {
            DropDownList.Items.Clear();
            ListItem ilkveri = new ListItem(ilkbaslik.ToString(), "", true); ilkveri.Selected = true;
            DropDownList.Items.Add(ilkveri);
            DropDownList.DataSource = DataTable;
            DropDownList.DataTextField = DataTable.Columns[1].ToString();
            DropDownList.DataValueField = DataTable.Columns[0].ToString();
            DropDownList.DataBind();
        }
        catch { }
    }
    //----------------------------------------------------------------------------------------------

    public static void DropDownListAyniBos(DropDownList DropDownList, String ilkbaslik, DataTable DataTable)
    {
        try
        {
            DropDownList.Items.Clear();
            ListItem ilkveri = new ListItem(ilkbaslik.ToString(), "", true); ilkveri.Selected = true;
            DropDownList.Items.Add(ilkveri);
            DropDownList.DataSource = DataTable;
            DropDownList.DataTextField = DataTable.Columns[1].ToString();
            DropDownList.DataValueField = DataTable.Columns[1].ToString();
            DropDownList.DataBind();
        }
        catch { }
    }
    //----------------------------------------------------------------------------------------------

    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------
    //---------------------------------   Meta Tag Fonksiyonları  ------------- --------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------

    public static string SetTitle(string metaContent, Page p)
    {
        string SiteBaslik = "";
        try
        {
            if (!String.IsNullOrEmpty(metaContent))
            { p.Title = Ayarlar.GoogleBaslik + metaContent; }
        }
        catch { }
        return SiteBaslik;
    }
    //----------------------------------------------------------------------------------------------

    public static string SetMeta(object metaContent, Page p, string MetaType)
    {
        string SiteHeadB = "";
        try
        {
            string _s = metaContent.ToString();
            if (!String.IsNullOrEmpty(_s))
            {
                HtmlMeta description = new HtmlMeta();
                description.Name = MetaType;
                description.Content = _s;
                p.Header.Controls.Add(description);
            }
        }
        catch { }
        return SiteHeadB;
    }
    //----------------------------------------------------------------------------------------------

    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------- Tarih Fonksiyonlar-------------------- -------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------

    public static string TarihFormatiniTersCevir(string gelenveri)
    {
        string temp = null;
        try
        {
            if (gelenveri == null)
            { temp = "01.01.2015"; }
            else { temp = (gelenveri); }
            string gun = temp.Substring(0, 2);
            string ay = temp.Substring(3, 2);
            string yil = temp.Substring(6, 4);
            if (gun.ToString().Length == 1)
            { gun = "0" + gun; }
            if (ay.ToString().Length == 1)
            { ay = "0" + ay; }
            temp = yil + ay + gun;
        }
        catch { }
        return temp;
    }
    //----------------------------------------------------------------------------------------------

    public static string TarihFormatiDuzCevir(string gelenveri)
    {
        string temp = null;
        try
        {
            if (gelenveri == null)
            { temp = "20150101"; }
            else
            { temp = System.Convert.ToString(gelenveri); }
            temp = temp.Substring(6, 2) + "." + temp.Substring(4, 2) + "." + temp.Substring(0, 4);
        }
        catch { }
        return temp;
    }
    ///----------------------------------------------------------------------------------------------

    public static string SaatFormatiTersCevir(string gelenveri)
    {
        string temp = null;
        try
        {
            if (gelenveri == null)
            { temp = "00:00"; }
            else { temp = (gelenveri); }
            string saat = temp.Substring(0, 2);
            string dakika = temp.Substring(3, 2);
            if (saat.ToString().Length == 1)
            { saat = "0" + saat; }
            if (dakika.ToString().Length == 1)
            { dakika = "0" + dakika; }
            temp = dakika + saat;
        }
        catch { }
        return temp;
    }
    //----------------------------------------------------------------------------------------------

    public static string SaatFormatiDuzCevir(string gelenveri)
    {
        string temp = null;
        try
        {
            if (gelenveri == null)
            { temp = "0000"; }
            else
            { temp = System.Convert.ToString(gelenveri); }
            temp = temp.Substring(2, 2) + ":" + temp.Substring(0, 2);
        }
        catch { }
        return temp;
    }
    ///----------------------------------------------------------------------------------------------

    public static int GunFarkikBul(DateTime dt1, DateTime dt2)
    {
        try
        {
            TimeSpan ts = dt1.Subtract(dt2);
            int gunSayisi = Convert.ToInt32(ts.TotalDays);
            return gunSayisi;
        }
        catch { return 0; }
    }
    ///----------------------------------------------------------------------------------------------

    public static int SaatFarkikBul(DateTime dt1, DateTime dt2)
    {
        try
        {
            TimeSpan ts = dt1.Subtract(dt2);
            int gunSayisi = Convert.ToInt32(ts.TotalMinutes);
            return gunSayisi;
        }
        catch { return 0; }
    }
    ///----------------------------------------------------------------------------------------------

    public static string DakikaFarkiBul(DateTime AlisTarihi, DateTime IadeTarihi)
    {
        TimeSpan tS = IadeTarihi - AlisTarihi; // iki zamanın farkını alır
        int dakika = (int)tS.TotalMinutes; // fark dakika cinsinden
        int saat = (int)tS.TotalHours; // fark saat cinsinden
        int gun = (int)tS.TotalDays; // fark gün cinsinden
        return saat.ToString();
    }
    ///----------------------------------------------------------------------------------------------
    
    public static string TarihEkli(string gelenveri)
    { return DateTime.Now.AddDays(Convert.ToDouble(gelenveri)).ToShortDateString().ToString(); }
    ///----------------------------------------------------------------------------------------------

    public static string EnSonKayitTarihi
    { get { return Fonksiyon.TarihFormatiniTersCevir(DateTime.Now.ToString().Substring(0, 10)) + DateTime.Now.ToString().Substring(11, 8).Replace(":", ""); } }
    ///----------------------------------------------------------------------------------------------

    public static string GunTamTarihSec(string tarih)
    {
        try
        { return Convert.ToDateTime(tarih).ToString("D"); }
        catch { return null; }
    }
    ///----------------------------------------------------------------------------------------------

    public static string TarihSec(string tarih)
    {
        try
        { return Convert.ToDateTime(tarih).ToString("dd.MM.yyyy"); }
        catch { return null; }
    }
    ///----------------------------------------------------------------------------------------------
    
    public static string GunTarihSec(string tarih)
    {
        try
        { return Convert.ToDateTime(tarih).ToString("dd MMMM yyyy"); }
        catch { return null; }
    }
    ///----------------------------------------------------------------------------------------------

    public static string GunTarihTamSec(string tarih)
    {
        try
        { return Convert.ToDateTime(tarih).ToString("dd MMMM yyyy, dddd"); }
        catch { return null; }
    }
    ///----------------------------------------------------------------------------------------------

    public static string GununTarihi
    { get { return DateTime.Now.ToString("dd MMMM yyyy, dddd"); } }
    ///----------------------------------------------------------------------------------------------

    public static string GununSaati
    { get { return DateTime.Now.ToString("hh:mm"); } }
    ///----------------------------------------------------------------------------------------------

	public static string SekilGunTarihSec(string tarih)
    {
        try
        { return Convert.ToDateTime(tarih).ToString("<b>dd<&b><br> MMMM<br>yyyy").Replace("<&b>", "</b>"); }
        catch { return null; }
    }
    ///----------------------------------------------------------------------------------------------

    public static string TarihFormati(string gelenveri)
    {
        //------------------------------------------------------------
        string veri1 = "", veri2 = "", veri3 = "";
        //------------------------------------------------------------
        string geridonenveri = gelenveri;
        //------------------------------------------------------------
        try
        {
            //------------------------------------------------------------
            string s = gelenveri;
            //------------------------------------------------------------
            string[] words = s.Split('.');
            for (int i = 0; i < words.Length; i++)
            {
                if (i == 0)
                {
                    if (words[0].Length == 1)
                    { veri1 = "0" + words[0]; }
                    else
                    { veri1 = words[0]; }
                }
                else if (i == 1)
                {
                    if (words[1].Length == 1)
                    { veri2 = "0" + words[1]; }
                    else
                    { veri2 = words[1]; }
                }
                else if (i == 2)
                {
                    veri3 = Fonksiyon.KarakterKisitlamaSec(words[2], 4);
                }
            }
            //------------------------------------------------------------
            geridonenveri = veri1 + "." + veri2 + "." + veri3;
            //------------------------------------------------------------
        }
        catch { }
        //------------------------------------------------------------
        return geridonenveri;
        //------------------------------------------------------------
    }
    //----------------------------------------------------------------------------------------------

    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------- Para Fonksiyonlar-------------------- -------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------

    public static decimal VirgulluTutarSec(decimal tutar, int virgulsonrasisayi)
    {
        try
        {
            string[] dec = tutar.ToString().Split(',');
            if (dec.Length > 0)
            {
                if (dec[1].Length > virgulsonrasisayi)
                { tutar = Convert.ToDecimal(dec[0] + "," + dec[1].Substring(0, virgulsonrasisayi)); }
            }
            return tutar;
        }
        catch { return 0; }
    }
    //----------------------------------------------------------------------------------------------

    public static string TutarVeri(decimal tutar)
    {
        //1234.560
        //return String.Format("{0:0.###}", tutar); //1234,56
        //return String.Format("{0:0.000}", tutar); //1234,560
        //return String.Format("{0:c}", tutar); //1.234,56 TL
        //return String.Format("{0:00.0}", tutar); //1234,6
        //return String.Format("{0:0,0.0}", tutar); //1.234,6
        //return String.Format("{0:0,0}", tutar); //1.235
        //return String.Format("{0:0}", tutar); //1.235
        //return String.Format("{0:0%}", tutar); //123456
        return String.Format("{0:n}", tutar); //1.234,56
        //return String.Format("{0:g}", tutar); //1234,56
    }
    //----------------------------------------------------------------------------------------------

    private string TutarYaziyaCevir(decimal tutar)
    {
        try
        {
            string sTutar = tutar.ToString("F2").Replace('.', ','); // Replace('.',',') ondalık ayracının . olma durumu için            
            string lira = sTutar.Substring(0, sTutar.IndexOf(',')); //tutarın tam kısmı
            string kurus = sTutar.Substring(sTutar.IndexOf(',') + 1, 2);
            string yazi = "";
            //----------------------------------------------------------------------------------------------------------------
            string[] birler = { "", "BİR ", "İKİ ", "ÜÇ ", "DÖRT ", "BEŞ ", "ALTI ", "YEDİ ", "SEKİZ ", "DOKUZ " };
            string[] onlar = { "", "ON ", "YİRMİ ", "OTUZ ", "KIRK ", "ELLİ ", "ALTMIŞ ", "YETMİŞ ", "SEKSEN ", "DOKSAN " };
            string[] binler = { "KATRİLYON ", "TRİLYON ", "MİLYAR ", "MİLYON ", "BİN ", "" }; //KATRİLYON'un önüne ekleme yapılarak artırabilir.
            //----------------------------------------------------------------------------------------------------------------
            int grupSayisi = 6; //sayıdaki 3'lü grup sayısı. katrilyon içi 6. (1.234,00 daki grup sayısı 2'dir.)
            //KATRİLYON'un başına ekleyeceğiniz her değer için grup sayısını artırınız.
            //----------------------------------------------------------------------------------------------------------------
            lira = lira.PadLeft(grupSayisi * 3, '0'); //sayının soluna '0' eklenerek sayı 'grup sayısı x 3' basakmaklı yapılıyor.            
            //----------------------------------------------------------------------------------------------------------------
            string grupDegeri;
            //----------------------------------------------------------------------------------------------------------------
            for (int i = 0; i < grupSayisi * 3; i += 3) //sayı 3'erli gruplar halinde ele alınıyor.
            {
                grupDegeri = "";
                if (lira.Substring(i, 1) != "0")
                    grupDegeri += birler[Convert.ToInt32(lira.Substring(i, 1))] + "YÜZ "; //yüzler                

                if (grupDegeri == " BİRYÜZ ") //biryüz düzeltiliyor.
                    grupDegeri = " YÜZ ";
                grupDegeri += onlar[Convert.ToInt32(lira.Substring(i + 1, 1))]; //onlar
                grupDegeri += birler[Convert.ToInt32(lira.Substring(i + 2, 1))]; //birler                
                if (grupDegeri != "") //binler
                    grupDegeri += binler[i / 3];
                if (grupDegeri == " BİRBİN ") //birbin düzeltiliyor.
                    grupDegeri = " BİN ";
                yazi += grupDegeri;
            }
            if (yazi != "")
                yazi += " TL ";
            int yaziUzunlugu = yazi.Length;
            if (kurus.Substring(0, 1) != "0") //kuruş onlar
                yazi += onlar[Convert.ToInt32(kurus.Substring(0, 1))];
            if (kurus.Substring(1, 1) != "0") //kuruş birler
                yazi += birler[Convert.ToInt32(kurus.Substring(1, 1))];
            if (yazi.Length > yaziUzunlugu)
                yazi += " KR.";
            else
                yazi += "SIFIR KR.";
            return yazi;
        }
        catch { return null; }
    }
    //----------------------------------------------------------------------------------------------

    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------- Genel Fonksiyonlar-------------------- -------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------

    public static string YandexTranslate(string from, string to, string text)
    {
        /* Access token talebinde bulun */
        HttpWebRequest req = (HttpWebRequest)WebRequest.Create("https://ceviri.yandex.net/api/v1/tr.json/translate?lang=" + from + "-" + to + "&text=" + text + "&srv=tr-text");
        req.Method = "GET";
        req.ContentType = "application/x-www-form-urlencoded";
        req.KeepAlive = true;
        HttpWebResponse res = (HttpWebResponse)req.GetResponse();
        StreamReader sr = new StreamReader(res.GetResponseStream());
        string source = sr.ReadToEnd();

        source = source.Replace("ya_.json.c(18)(", "").Replace(")", "");
        var results = JsonConvert.DeserializeObject<dynamic>(source);
        return (string)results.text[0];
    }
    //----------------------------------------------------------------------------------------------

    //----------------------------------------------------------------------------------------------
    public static string YaziyiResmeCevir(string yazi)
    {
        try
        {
            string text = yazi.Trim();
            Bitmap bitmap = new Bitmap(1, 1);
            Font font = new Font("Arial", 12, FontStyle.Regular, GraphicsUnit.Pixel);
            Graphics graphics = Graphics.FromImage(bitmap);
            int width = (int)graphics.MeasureString(text, font).Width;
            int height = (int)graphics.MeasureString(text, font).Height;
            bitmap = new Bitmap(bitmap, new Size(width, height));
            graphics = Graphics.FromImage(bitmap);
            graphics.Clear(Color.White);
            graphics.SmoothingMode = SmoothingMode.AntiAlias;
            graphics.TextRenderingHint = TextRenderingHint.AntiAlias;
            graphics.DrawString(text, font, new SolidBrush(Color.FromArgb(255, 0, 0)), 0, 0);
            graphics.Flush();
            graphics.Dispose();
            string fileName = Path.GetFileNameWithoutExtension(HttpContext.Current.Request.ServerVariables["URL"].Replace(".", "").Replace("/", "") + yazi) + ".png";
            bitmap.Save(HttpContext.Current.Server.MapPath(TempCropKlasoru) + fileName, ImageFormat.Png);
            return (Fonksiyon.TempCropKlasoru + fileName);
        }
        catch { return null; }
    }
    //----------------------------------------------------------------------------------------------

    public static string YaziSifrele(string veri)
    {
        // gelen veri byte dizisine aktarılıyor 
        byte[] veriByteDizisi = System.Text.ASCIIEncoding.ASCII.GetBytes(veri);
        // base64 şifreleme algoritmasına göre şifreleniyor.
        string sifrelenmisVeri = System.Convert.ToBase64String(veriByteDizisi);
        return sifrelenmisVeri;
    }
    //----------------------------------------------------------------------------------------------

    public static string YaziSifreCoz(string cozVeri)
    {
        byte[] cozByteDizi = System.Convert.FromBase64String(cozVeri);
        string orjinalVeri = System.Text.ASCIIEncoding.ASCII.GetString(cozByteDizi);
        return orjinalVeri;
    }
    //----------------------------------------------------------------------------------------------

    public static string Random()
    { return DateTime.Now.Day.ToString() + DateTime.Now.Month + DateTime.Now.Year + DateTime.Now.Hour + DateTime.Now.Second + DateTime.Now.Minute + DateTime.Now.Millisecond; }
    //----------------------------------------------------------------------------------------------

    public static string RandomSayi
    {
        get { return new Random().Next(1, 999999).ToString(); }
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimRandomSayi
    {
        get { return new Random().Next(10000, 99999).ToString(); }
    }
    //----------------------------------------------------------------------------------------------

    public static string UrlRandomSayi
    {
        get { return new Random().Next(1, 999).ToString(); }
    }
    //----------------------------------------------------------------------------------------------

    public static string md5sifreleme(string sifrele)
    { return System.Web.Security.FormsAuthentication.HashPasswordForStoringInConfigFile(sifrele, "md5"); }
    //----------------------------------------------------------------------------------------------

    public static bool EmailAdresiKontrol(string inputEmail)
    {
        try
        {
            const string strRegex = @"^([a-zA-Z0-9_\-\.]+)@((\[[0-9]{1,3}" +
              @"\.[0-9]{1,3}\.[0-9]{1,3}\.)|(([a-zA-Z0-9\-]+\" +
              @".)+))([a-zA-Z]{2,4}|[0-9]{1,3})(\]?)$";

            return (new Regex(strRegex)).IsMatch(inputEmail);
        }
        catch { return false; }
    }
    //----------------------------------------------------------------------------------------------

    public static string SonKelimeyiAlmak(string gelenveri)
    {
        try
        {
            //--------------------------------------------------------------------------------
            var gelenveriBol = gelenveri.Replace("  ", " ").Trim().Split(' ');
            //--------------------------------------------------------------------------------
            if (gelenveriBol.Count() > 0)
            { gelenveri = gelenveriBol[gelenveriBol.Count() - 1]; }
            //--------------------------------------------------------------------------------
            return gelenveri;
            //--------------------------------------------------------------------------------
        }
        catch { return null; }
    }
    //----------------------------------------------------------------------------------------------

    public static string SonKarakteriTemizle(string gelenveri)
    {
        try
        {
            string[] gelenveri2 = gelenveri.Split(new char[] { ',' }, StringSplitOptions.RemoveEmptyEntries);
            string sonuc = gelenveri2[0];
            for (int i = 1; i < gelenveri2.Length; i++)
            { sonuc += "," + gelenveri2[i]; }
            return sonuc;
        }
        catch { return null; }
    }
    //----------------------------------------------------------------------------------------------

    public static string KarakterKisitlamaSec(string gelenveri, int karakteruzunlugu)
    {
        try
        {
            if (gelenveri.Length > karakteruzunlugu)
                return gelenveri.Substring(0, karakteruzunlugu);
            return gelenveri;
        }
        catch { return null; }
    }
    //-----------------------------------------------------------------------------------------------

    public static string KarakterKisitlamaNoktaliSec(string gelenveri, int karakteruzunlugu)
    {
        try
        {
            if (gelenveri.Length > karakteruzunlugu)
                return gelenveri.Substring(0, karakteruzunlugu) + "...";
            return gelenveri;
        }
        catch
        { return null; }
    }
    //-----------------------------------------------------------------------------------------------

    public static bool KarakterKisitlamaKontrol(string gelenveri, int karakteruzunlugu)
    {
        bool durum = false;
        try
        {
            if (gelenveri.Length > karakteruzunlugu)
            { durum = true; }
        }
        catch { }
        return durum;
    }
    //-----------------------------------------------------------------------------------------------

    public static string IlkHarflerBuyuk(object text)
    {
        try
        { return System.Globalization.CultureInfo.CurrentCulture.TextInfo.ToTitleCase(text.ToString().ToLower(new System.Globalization.CultureInfo(SiteDilKodu(), false))); }
        catch { return null; }
    }
    //----------------------------------------------------------------------------------------------

    public static string TumHarflerBuyuk(object text)
    {
        try
        {
            return text.ToString().ToUpper(new System.Globalization.CultureInfo(SiteDilKodu(), false));
            //return text.ToString().ToUpper().Replace("İ","I");  // En İçin
            //return text.ToString().ToUpperInvariant();  // TR İçin
        }
        catch { return null; }
    }
    //----------------------------------------------------------------------------------------------

    public string IpAdresi
    { get { return System.Web.HttpContext.Current.Request.ServerVariables["REMOTE_HOST"]; } }
    //----------------------------------------------------------------------------------------------

    public static string SiteDilKodu()
    {
        string DilKodu = "tr-TR";
        if (Ayarlar.SiteDilDurum == true)
        {
            //----------------------------------------------------------------------------------------------
            if (!String.IsNullOrEmpty(Ayarlar.SiteDilKoduSec))
            { DilKodu = Ayarlar.SiteDilKoduSec; }
            else
            { DilKodu = Ayarlar.SiteDilKoduText; }
            //----------------------------------------------------------------------------------------------
            System.Threading.Thread.CurrentThread.CurrentUICulture = new System.Globalization.CultureInfo(DilKodu);
            System.Threading.Thread.CurrentThread.CurrentCulture = new System.Globalization.CultureInfo(DilKodu);
            //----------------------------------------------------------------------------------------------
        }
        return DilKodu;
    }
    //----------------------------------------------------------------------------------------------

    public static string HtmlTemizlemeSec(string gelenveri)
    {
        try
        {
            gelenveri = Regex.Replace(gelenveri, @"<(.|\n)*?>", string.Empty);
            gelenveri = gelenveri.Replace("&nbsp;", "");
            return gelenveri;
        }
        catch { return null; }
    }
    //-----------------------------------------------------------------------------------------------

    public static string RedirectYonlendirme(string simdikisite, string yonleneceksite)
    {
        if (HttpContext.Current.Request.Url.ToString().ToLower().Contains(simdikisite))
        {
            HttpContext.Current.Response.Clear();
            HttpContext.Current.Response.Status = "301 Moved Permanently";
            HttpContext.Current.Response.AddHeader("Location", HttpContext.Current.Request.Url.ToString().Replace(simdikisite, yonleneceksite));
        }
        return null;
    }
    //-----------------------------------------------------------------------------------------------

    public static bool SadeceSayiKontrol(string gelenveri)
    {
        bool durum = false;
        double oReturn = 0;
        try
        {
            var result = double.TryParse(gelenveri, out oReturn);
            if (result == true)
            { durum = true; }
        }
        catch {}
        return durum;
    }
    //----------------------------------------------------------------------------------------------

    public static bool BoslukKontrol(string gelenveri)
    {
        bool durum = false;
        if (gelenveri != null)
        {
            if (gelenveri.Trim().IndexOf(" ") != -1)
            { durum = true; }
        }
        return durum;
    }
    //----------------------------------------------------------------------------------------------

    public static string SayfaBulunamadi()
    {
        StringBuilder sb = new StringBuilder();
        sb.Append("<HTML><HEAD><TITLE>The page cannot be found</TITLE>");
        sb.Append("<META HTTP-EQUIV=\"Content-Type\" Content=\"text/html; charset=Windows-1252\">");
        sb.Append("<STYLE type=\"text/css\">");
        sb.Append("  BODY { font: 8pt/12pt verdana }");
        sb.Append("  H1 { font: 13pt/15pt verdana }");
        sb.Append("  H2 { font: 8pt/12pt verdana }");
        sb.Append("  A:link { color: red }");
        sb.Append("  A:visited { color: maroon }");
        sb.Append("</STYLE>");
        sb.Append("</HEAD><BODY><TABLE width=500 border=0 cellspacing=10><TR><TD>");
        sb.Append("<h1>The page cannot be found</h1>");
        sb.Append("The page you are looking for might have been removed, had its name changed, or ");
        sb.Append("is temporarily unavailable.");
        sb.Append("<hr>");
        sb.Append("<p>Please try the following:</p>");
        sb.Append("<ul>");
        sb.Append("<li>Make sure that the Web site address displayed in the address bar of your ");
        sb.Append("browser is spelled and formatted correctly.</li>");
        sb.Append("<li>If you reached this page by clicking a link, contact the Web site ");
        sb.Append("administrator to alert them that the link is incorrectly formatted. ");
        sb.Append("</li>");
        sb.Append("<li>Click the <a href=\"javascript:history.back(1)\">Back</a> button to try ");
        sb.Append("another link.</li>");
        sb.Append("</ul>");
        sb.Append("<h2>HTTP Error 404 - File or directory not found.<br>Internet Information ");
        sb.Append("Services (IIS)</h2>");
        sb.Append("<hr>");
        sb.Append("<p>Technical Information (for support personnel)</p>");
        sb.Append("<ul>");
        sb.Append("<li>Go to <a href=\"http://go.microsoft.com/fwlink/?linkid=8180\">Microsoft ");
        sb.Append("Product Support Services</a> and perform a title search for the words <b>HTTP</b> ");
        sb.Append("and <b>404</b>.</li>");
        sb.Append("<li>Open <b>IIS Help</b>, which is accessible in IIS Manager (inetmgr), and ");
        sb.Append("search for topics titled <b>Web Site Setup</b>, <b>Common Administrative Tasks</b>,");
        sb.Append("and <b>About Custom Error Messages</b>.</li>");
        sb.Append("</ul>");
        sb.Append("</TD></TR></TABLE></BODY></HTML>");
        return sb.ToString();
        HttpContext.Current.Response.End();
    }
    //----------------------------------------------------------------------------------------------

    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------- Klasör Adresleri-------------------- -------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------

    public static string ResolveUrl(string relativeUrl)
    { System.Web.UI.Page p = HttpContext.Current.Handler as System.Web.UI.Page; return p.ResolveUrl(relativeUrl); }
    //----------------------------------------------------------------------------------------------

    public static string SifreliResimDosya(string resimadi)
    {
        string imgsrc = "";
        try
        {
            string base64Verisi = Convert.ToBase64String(System.IO.File.ReadAllBytes(HttpContext.Current.Server.MapPath("Uploads/GenelResim/" + resimadi)));
            imgsrc = "data:image/png;base64," + base64Verisi + "";
        }
        catch { }
        return imgsrc;
    }
    //----------------------------------------------------------------------------------------------

    public static string SiteResimKlasoru
    {
        get
        {
            string klasor = "";
            if (Ayarlar.WatermakDurum == true && !String.IsNullOrEmpty(Ayarlar.WatermakResim))
            { klasor = @"" + ResolveUrl("~") + "WaterMark.aspx?img=/Uploads/GenelResim/"; }
            else
            { klasor = @"" + ResolveUrl("~") + "Uploads/GenelResim/"; }
            return klasor;
        }
    }
    //----------------------------------------------------------------------------------------------
    public static string BannerKlasoru { get { return @"" + ResolveUrl("~") + "Uploads/GenelResim/"; } }
    //----------------------------------------------------------------------------------------------

    public static string KucukResimKlasoru
    {
        get
        {
            string klasor = "";
            if (Ayarlar.KucukResimWatermakDurum == true && Ayarlar.WatermakDurum == true && !String.IsNullOrEmpty(Ayarlar.WatermakResim))
            { klasor = @"" + ResolveUrl("~") + "WaterMark.aspx?kz=1&img=/Uploads/GenelResim/"; }
            else
            { klasor = @"" + ResolveUrl("~") + "Uploads/GenelResim/"; }
            return klasor;
        }
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimKlasoru { get { return @"" + ResolveUrl("~") + "Uploads/GenelResim/"; } }
    //----------------------------------------------------------------------------------------------

    public static string DosyaKlasoru { get { return @"" + ResolveUrl("~") + "Uploads/GenelDosya/"; } }
    //----------------------------------------------------------------------------------------------

    public static string VideoKlasoru { get { return @"" + ResolveUrl("~") + "Uploads/GenelVideo/"; } }
    //----------------------------------------------------------------------------------------------

    public static string ResimKlasoruPanelMenu { get { return @"" + ResolveUrl("~") + "Uploads/PanelMenu/"; } }
    //----------------------------------------------------------------------------------------------

    public static string ResimKlasoruEditorResim { get { return @"" + ResolveUrl("~") + "Uploads/EditorResim/"; } }
    //----------------------------------------------------------------------------------------------

    public static string ResimKlasoruEditorDosya { get { return @"" + ResolveUrl("~") + "Uploads/EditorDosya/"; } }
    //----------------------------------------------------------------------------------------------

    public static string WatermakResimKlasoru { get { return @"" + ResolveUrl("~") + "WaterMark.aspx?img=/Uploads/GenelResim/"; } }
    //----------------------------------------------------------------------------------------------

    public static string TempKlasoru { get { return @"" + ResolveUrl("~") + "Uploads/GenelTemp/"; } }
    //----------------------------------------------------------------------------------------------

    public static string TempCropKlasoru { get { return @"" + ResolveUrl("~") + "Uploads/GenelTemp/CropTemp/"; } }
    //----------------------------------------------------------------------------------------------


    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------
    //--------------------------- Resim ve Dosya Yükleme -------------------------------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------

    public static Bitmap ResimCiz(Bitmap Resim, Bitmap CizilecekResim)
    {
        int ResimWidth = Resim.Width;
        int ResimHeight = Resim.Height;
        int CizilecekResimWidth = CizilecekResim.Width;
        int CizilecekResimHeight = CizilecekResim.Height;

        Graphics g = Graphics.FromImage(Resim);
        g.SmoothingMode = SmoothingMode.HighQuality;
        if (Ayarlar.ResimUzerineLogoTur == 0)
        { g.DrawImage(CizilecekResim, new Point() { X = (ResimWidth - CizilecekResimWidth) / 2, Y = (ResimHeight - CizilecekResimHeight) / 2 }); }
        else
        { g.DrawImage(CizilecekResim, new Point() { X = Ayarlar.ResimUzerineLogoSolBosluk, Y = Ayarlar.ResimUzerineLogoYukariBosluk }); }
        return Resim;
    }
    //----------------------------------------------------------------------------------------------

    public static System.Drawing.Image ResimBoyutlandirGenislik(System.Drawing.Image imgPhoto, int yukseklik)
    {
        int sourceWidth = imgPhoto.Width;
        int sourceHeight = imgPhoto.Height;

        int destWidth = yukseklik;
        int destHeight = sourceHeight * yukseklik / imgPhoto.Width; //Resmin bozulmaması için en boy ayarını veriyoruz.

        Bitmap bmPhoto = new Bitmap(destWidth, destHeight);
        bmPhoto.SetResolution(imgPhoto.HorizontalResolution, imgPhoto.VerticalResolution);

        Graphics grPhoto = Graphics.FromImage(bmPhoto);
        grPhoto.InterpolationMode = InterpolationMode.HighQualityBicubic; // Resmin kalitesini ayarlıyoruz.
        grPhoto.FillRectangle(Brushes.Transparent, 0, 0, destWidth, destHeight);

        grPhoto.DrawImage(imgPhoto, new Rectangle(0, 0, destWidth, destHeight), new Rectangle(0, 0, sourceWidth, sourceHeight), GraphicsUnit.Pixel);

        grPhoto.Dispose();
        return bmPhoto;
    }
    //----------------------------------------------------------------------------------------------

    static System.Drawing.Image ResimBoyutlandirYukseklik(System.Drawing.Image imgPhoto, int genislik)
    {
        int sourceWidth = imgPhoto.Width;
        int sourceHeight = imgPhoto.Height;

        int destHeight = genislik;
        int destWidth = sourceWidth * genislik / imgPhoto.Height; //Resmin bozulmaması için en boy ayarını veriyoruz.

        Bitmap bmPhoto = new Bitmap(destWidth, destHeight);
        bmPhoto.SetResolution(imgPhoto.HorizontalResolution, imgPhoto.VerticalResolution);

        Graphics grPhoto = Graphics.FromImage(bmPhoto);
        grPhoto.InterpolationMode = InterpolationMode.HighQualityBicubic; // Resmin kalitesini ayarlıyoruz.
        grPhoto.FillRectangle(Brushes.Transparent, 0, 0, destWidth, destHeight);

        grPhoto.DrawImage(imgPhoto, new Rectangle(0, 0, destWidth, destHeight), new Rectangle(0, 0, sourceWidth, sourceHeight), GraphicsUnit.Pixel);

        grPhoto.Dispose();
        return bmPhoto;
    }
    //----------------------------------------------------------------------------------------------

    static System.Drawing.Image OzelResimBoyutlandir(System.Drawing.Image imgPhoto, int yukseklik, int genislik)
    {
        int sourceWidth = imgPhoto.Width;
        int sourceHeight = imgPhoto.Height;

        int destWidth = genislik;
        int destHeight = yukseklik;

        Bitmap bmPhoto = new Bitmap(destWidth, destHeight);
        bmPhoto.SetResolution(imgPhoto.HorizontalResolution, imgPhoto.VerticalResolution);

        Graphics grPhoto = Graphics.FromImage(bmPhoto);
        grPhoto.InterpolationMode = InterpolationMode.HighQualityBicubic; // Resmin kalitesini ayarlıyoruz.
        grPhoto.FillRectangle(Brushes.Transparent, 0, 0, destWidth, destHeight);

        grPhoto.DrawImage(imgPhoto, new Rectangle(0, 0, destWidth, destHeight), new Rectangle(0, 0, sourceWidth, sourceHeight), GraphicsUnit.Pixel);

        grPhoto.Dispose();
        return bmPhoto;
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimDirekYukleme(HttpPostedFile GozatYolu, string KlasorAdi, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png" | uzanti == ".swf" | uzanti == ".rar" | uzanti == ".zip" | uzanti == ".pdf" | uzanti == ".doc" | uzanti == ".docx" | uzanti == ".xls" | uzanti == ".xlsx" | uzanti == ".pps" | uzanti == ".ppsx" | uzanti == ".ppt" | uzanti == ".pptx" | uzanti == ".ico") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi; string random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------

            string tamYol = HttpContext.Current.Server.MapPath(KlasorAdi + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //---------------------------------------------------------------------------------------------------
            if (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png")
            {
                //-----------------------------------------------------------------------------------------------------------------------------------
                System.Drawing.Image imgPhotoVert1 = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
                System.Drawing.Image imgPhoto = null;
                imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert1, imgPhotoVert1.Width);
                imgPhotoVert1.Dispose();
                //-----------------------------------------------------------------------------------------------------------------------------------
                if (Ayarlar.TekliResimKaliteDurum == false)
                {
                    if (uzanti == ".jpg" | uzanti == ".jpeg")
                    { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                    else if (uzanti == ".gif")
                    { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                    else if (uzanti == ".png")
                    { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                    else
                    { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                }
                else
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                //-----------------------------------------------------------------------------------------------------------------------------------
                imgPhoto.Dispose();
                //-----------------------------------------------------------------------------------------------------------------------------------
            }
            //-----------------------------------------------------------------------------------------------------------------------------------
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {

                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimGenislikTekliKesme(HttpPostedFile GozatYolu, string KlasorAdi, int Genislik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //-----------------------------------------------------------------------------------------------------------------------------------
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            //-----------------------------------------------------------------------------------------------------------------------------------
            System.Drawing.Image imgPhotoVert = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto = null;
            if (imgPhotoVert.Width > Genislik)
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, Genislik); }
            else
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, imgPhotoVert.Width); }
            imgPhotoVert.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.BuyukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            new FileInfo(tamYol).Delete();
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimGenislikCiftliKesme(HttpPostedFile GozatYolu, string KlasorAdi, int KucukResimGenislik, int BuyukResimGenislik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //-----------------------------------------------------------------------------------------------------------------------------------
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            //-----------------------------------------------------------------------------------------------------------------------------------

            System.Drawing.Image imgPhotoVert = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto = null;
            if (imgPhotoVert.Width > KucukResimGenislik)
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, KucukResimGenislik); }
            else
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, imgPhotoVert.Width); }
            imgPhotoVert.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------

            if (Ayarlar.KucukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }

            //-------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Dispose();

            //-----------------------------------------------------------------------------------------------------------------------------------
            System.Drawing.Image imgPhotoVert1 = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto1 = null;
            if (imgPhotoVert1.Width > BuyukResimGenislik)
            { imgPhoto1 = ResimBoyutlandirGenislik(imgPhotoVert1, BuyukResimGenislik); }
            else
            { imgPhoto1 = ResimBoyutlandirGenislik(imgPhotoVert1, imgPhotoVert1.Width); }
            imgPhotoVert1.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------

            if (Ayarlar.BuyukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }

            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto1.Dispose();

            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.ResimLogoDurum == true && !String.IsNullOrEmpty(Ayarlar.ResimLogo))
            {
                Bitmap bmp = new Bitmap(HttpContext.Current.Server.MapPath(resimKlasor + "/b_" + ResimAdi));
                Bitmap logo = new Bitmap(HttpContext.Current.Server.MapPath("/Uploads/GenelResim/" + Ayarlar.ResimLogo));
                Fonksiyon.ResimCiz(bmp, logo).Save(HttpContext.Current.Server.MapPath(resimKlasor + "/L_" + ResimAdi));
            }
            //-----------------------------------------------------------------------------------------------------------------------------------

            new FileInfo(tamYol).Delete();
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimGenislikUcluKesme(HttpPostedFile GozatYolu, string KlasorAdi, int KucukResimGenislik, int OrtaResimGenislik, int BuyukResimGenislik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //-----------------------------------------------------------------------------------------------------------------------------------
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            //-----------------------------------------------------------------------------------------------------------------------------------
            System.Drawing.Image imgPhotoVert = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto = null;
            if (imgPhotoVert.Width > KucukResimGenislik)
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, KucukResimGenislik); }
            else
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, imgPhotoVert.Width); }
            imgPhotoVert.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.KucukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------

            System.Drawing.Image imgPhotoVert1 = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto1 = null;
            if (imgPhotoVert1.Width > OrtaResimGenislik)
            { imgPhoto1 = ResimBoyutlandirGenislik(imgPhotoVert1, OrtaResimGenislik); }
            else
            { imgPhoto1 = ResimBoyutlandirGenislik(imgPhotoVert1, imgPhotoVert1.Width); }
            imgPhotoVert1.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.OrtaResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "o_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "o_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "o_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "o_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "o_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto1.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------

            System.Drawing.Image imgPhotoVert2 = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto2 = null;
            if (imgPhotoVert2.Width > BuyukResimGenislik)
            { imgPhoto2 = ResimBoyutlandirGenislik(imgPhotoVert2, BuyukResimGenislik); }
            else
            { imgPhoto2 = ResimBoyutlandirGenislik(imgPhotoVert2, imgPhotoVert2.Width); }
            imgPhotoVert2.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.BuyukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto2.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto2.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto2.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto2.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto2.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto2.Dispose();

            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.ResimLogoDurum == true && !String.IsNullOrEmpty(Ayarlar.ResimLogo))
            {
                Bitmap bmp = new Bitmap(HttpContext.Current.Server.MapPath(resimKlasor + "/b_" + ResimAdi));
                Bitmap logo = new Bitmap(HttpContext.Current.Server.MapPath(ResimKlasoru + Ayarlar.ResimLogo));
                Fonksiyon.ResimCiz(bmp, logo).Save(HttpContext.Current.Server.MapPath(resimKlasor + "/L_" + ResimAdi));
            }
            //-----------------------------------------------------------------------------------------------------------------------------------

            new FileInfo(tamYol).Delete();
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //-----------------------------------------------------------------------

    public static string ResimGenislikBannerKesme(HttpPostedFile GozatYolu, string KlasorAdi, int KucukResimGenislik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //-----------------------------------------------------------------------------------------------------------------------------------
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            //-----------------------------------------------------------------------------------------------------------------------------------

            System.Drawing.Image imgPhotoVert = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto = null;
            if (imgPhotoVert.Width > KucukResimGenislik)
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, KucukResimGenislik); }
            else
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, imgPhotoVert.Width); }
            imgPhotoVert.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.KucukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Dispose();

            //-----------------------------------------------------------------------------------------------------------------------------------
            System.Drawing.Image imgPhotoVert1 = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto1 = null;
            imgPhoto1 = ResimBoyutlandirGenislik(imgPhotoVert1, imgPhotoVert1.Width);
            imgPhotoVert1.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.TekliResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto1.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------

            //new FileInfo(tamYol).Delete();
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimYukseklikTekliKesme(HttpPostedFile GozatYolu, string KlasorAdi, int Yukseklik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //-----------------------------------------------------------------------------------------------------------------------------------
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            //-----------------------------------------------------------------------------------------------------------------------------------
            System.Drawing.Image imgPhotoVert = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto = null;
            if (imgPhotoVert.Height > Yukseklik)
            { imgPhoto = ResimBoyutlandirYukseklik(imgPhotoVert, Yukseklik); }
            else
            { imgPhoto = ResimBoyutlandirYukseklik(imgPhotoVert, imgPhotoVert.Height); }
            imgPhotoVert.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.BuyukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            new FileInfo(tamYol).Delete();
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimYukseklikCiftliKesme(HttpPostedFile GozatYolu, string KlasorAdi, int KucukResimYukseklik, int BuyukResimYukseklik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //-----------------------------------------------------------------------------------------------------------------------------------
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            //-----------------------------------------------------------------------------------------------------------------------------------

            System.Drawing.Image imgPhotoVert = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto = null;
            if (imgPhotoVert.Height > KucukResimYukseklik)
            { imgPhoto = ResimBoyutlandirYukseklik(imgPhotoVert, KucukResimYukseklik); }
            else
            { imgPhoto = ResimBoyutlandirYukseklik(imgPhotoVert, imgPhotoVert.Height); }
            imgPhotoVert.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.KucukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------

            System.Drawing.Image imgPhotoVert1 = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto1 = null;
            if (imgPhotoVert1.Height > BuyukResimYukseklik)
            { imgPhoto1 = ResimBoyutlandirYukseklik(imgPhotoVert1, BuyukResimYukseklik); }
            else
            { imgPhoto1 = ResimBoyutlandirYukseklik(imgPhotoVert1, imgPhotoVert1.Height); }
            imgPhotoVert1.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.BuyukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto1.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------

            if (Ayarlar.ResimLogoDurum == true && !String.IsNullOrEmpty(Ayarlar.ResimLogo))
            {
                Bitmap bmp = new Bitmap(HttpContext.Current.Server.MapPath(resimKlasor + "/b_" + ResimAdi));
                Bitmap logo = new Bitmap(HttpContext.Current.Server.MapPath(ResimKlasoru + Ayarlar.ResimLogo));
                Fonksiyon.ResimCiz(bmp, logo).Save(HttpContext.Current.Server.MapPath(resimKlasor + "/L_" + ResimAdi));
            }
            //-----------------------------------------------------------------------------------------------------------------------------------

            new FileInfo(tamYol).Delete();
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimYukseklikUcluKesme(HttpPostedFile GozatYolu, string KlasorAdi, int KucukResimYukseklik, int OrtaResimYukseklik, int BuyukResimYukseklik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //-----------------------------------------------------------------------------------------------------------------------------------
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            //-----------------------------------------------------------------------------------------------------------------------------------
            System.Drawing.Image imgPhotoVert = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto = null;
            if (imgPhotoVert.Height > KucukResimYukseklik)
            { imgPhoto = ResimBoyutlandirYukseklik(imgPhotoVert, KucukResimYukseklik); }
            else
            { imgPhoto = ResimBoyutlandirYukseklik(imgPhotoVert, imgPhotoVert.Height); }
            imgPhotoVert.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.KucukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Dispose();

            //-----------------------------------------------------------------------------------------------------------------------------------
            System.Drawing.Image imgPhotoVert1 = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto1 = null;
            if (imgPhotoVert1.Height > OrtaResimYukseklik)
            { imgPhoto1 = ResimBoyutlandirYukseklik(imgPhotoVert1, OrtaResimYukseklik); }
            else
            { imgPhoto1 = ResimBoyutlandirYukseklik(imgPhotoVert1, imgPhotoVert1.Height); }
            imgPhotoVert1.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.OrtaResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "o_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "o_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "o_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "o_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "o_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto1.Dispose();

            //-----------------------------------------------------------------------------------------------------------------------------------
            System.Drawing.Image imgPhotoVert2 = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto2 = null;
            if (imgPhotoVert2.Height > BuyukResimYukseklik)
            { imgPhoto2 = ResimBoyutlandirYukseklik(imgPhotoVert2, BuyukResimYukseklik); }
            else
            { imgPhoto2 = ResimBoyutlandirYukseklik(imgPhotoVert2, imgPhotoVert2.Height); }
            imgPhotoVert2.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.BuyukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto2.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto2.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto2.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto2.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto2.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto2.Dispose();

            //-----------------------------------------------------------------------------------------------------------------------------------

            if (Ayarlar.ResimLogoDurum == true && !String.IsNullOrEmpty(Ayarlar.ResimLogo))
            {
                Bitmap bmp = new Bitmap(HttpContext.Current.Server.MapPath(resimKlasor + "/b_" + ResimAdi));
                Bitmap logo = new Bitmap(HttpContext.Current.Server.MapPath(ResimKlasoru + Ayarlar.ResimLogo));
                Fonksiyon.ResimCiz(bmp, logo).Save(HttpContext.Current.Server.MapPath(resimKlasor + "/L_" + ResimAdi));
            }
            //-----------------------------------------------------------------------------------------------------------------------------------

            new FileInfo(tamYol).Delete();
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimYukseklikBannerKesme(HttpPostedFile GozatYolu, string KlasorAdi, int KucukResimYukseklik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //-----------------------------------------------------------------------------------------------------------------------------------
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            //-----------------------------------------------------------------------------------------------------------------------------------

            System.Drawing.Image imgPhotoVert = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto = null;
            if (imgPhotoVert.Height > KucukResimYukseklik)
            { imgPhoto = ResimBoyutlandirYukseklik(imgPhotoVert, KucukResimYukseklik); }
            else
            { imgPhoto = ResimBoyutlandirYukseklik(imgPhotoVert, imgPhotoVert.Height); }
            imgPhotoVert.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.KucukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------

            System.Drawing.Image imgPhotoVert1 = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto1 = null;
            imgPhoto1 = ResimBoyutlandirYukseklik(imgPhotoVert1, imgPhotoVert1.Height);
            imgPhotoVert1.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.BuyukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto1.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------

            if (Ayarlar.ResimLogoDurum == true && !String.IsNullOrEmpty(Ayarlar.ResimLogo))
            {
                Bitmap bmp = new Bitmap(HttpContext.Current.Server.MapPath(resimKlasor + "/b_" + ResimAdi));
                Bitmap logo = new Bitmap(HttpContext.Current.Server.MapPath(ResimKlasoru + Ayarlar.ResimLogo));
                Fonksiyon.ResimCiz(bmp, logo).Save(HttpContext.Current.Server.MapPath(resimKlasor + "/L_" + ResimAdi));
            }
            //-----------------------------------------------------------------------------------------------------------------------------------

            new FileInfo(tamYol).Delete();
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimOzelTekliKesme(HttpPostedFile GozatYolu, string KlasorAdi, int Genislik, int Yukseklik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //-----------------------------------------------------------------------------------------------------------------------------------
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            //-----------------------------------------------------------------------------------------------------------------------------------
            System.Drawing.Image imgPhotoVert = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto = null;
            if (imgPhotoVert.Width > Genislik && imgPhotoVert.Height > Yukseklik)
            { imgPhoto = OzelResimBoyutlandir(imgPhotoVert, Yukseklik, Genislik); }
            else
            { imgPhoto = OzelResimBoyutlandir(imgPhotoVert, imgPhotoVert.Height, imgPhotoVert.Width); }
            imgPhotoVert.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.BuyukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Dispose();

            //-----------------------------------------------------------------------------------------------------------------------------------
            new FileInfo(tamYol).Delete();
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimOzelCiftliKesme(HttpPostedFile GozatYolu, string KlasorAdi, int KucukGenislik, int KucukYukseklik, int BuyukGenislik, int BuyukYukseklik, string SiteResimAdi)
    {

        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //-----------------------------------------------------------------------------------------------------------------------------------
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            //-----------------------------------------------------------------------------------------------------------------------------------

            System.Drawing.Image imgPhotoVert = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto = null;
            if (imgPhotoVert.Width > KucukGenislik && imgPhotoVert.Height > KucukYukseklik)
            { imgPhoto = OzelResimBoyutlandir(imgPhotoVert, KucukYukseklik, KucukGenislik); }
            else
            { imgPhoto = OzelResimBoyutlandir(imgPhotoVert, imgPhotoVert.Height, imgPhotoVert.Width); }
            imgPhotoVert.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.KucukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------

            System.Drawing.Image imgPhotoVert1 = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto1 = null;
            if (imgPhotoVert.Width > BuyukGenislik && imgPhotoVert.Height > KucukYukseklik)
            { imgPhoto1 = OzelResimBoyutlandir(imgPhotoVert1, BuyukYukseklik, BuyukGenislik); }
            else
            { imgPhoto1 = OzelResimBoyutlandir(imgPhotoVert1, imgPhotoVert.Height, imgPhotoVert.Width); }
            imgPhotoVert1.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.BuyukResimKaliteDurum == false)
            {
                if (uzanti == ".jpg" | uzanti == ".jpeg")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
                else if (uzanti == ".gif")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
                else if (uzanti == ".png")
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
                else
                { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            }
            else
            { imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto1.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------

            if (Ayarlar.ResimLogoDurum == true && !String.IsNullOrEmpty(Ayarlar.ResimLogo))
            {
                Bitmap bmp = new Bitmap(HttpContext.Current.Server.MapPath(resimKlasor + "/b_" + ResimAdi));
                Bitmap logo = new Bitmap(HttpContext.Current.Server.MapPath(ResimKlasoru + Ayarlar.ResimLogo));
                Fonksiyon.ResimCiz(bmp, logo).Save(HttpContext.Current.Server.MapPath(resimKlasor + "/L_" + ResimAdi));
            }
            //-----------------------------------------------------------------------------------------------------------------------------------

            new FileInfo(tamYol).Delete();
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------



    //----------------------------------------------------------------------------------------------
    //-----------------------------------Resim Kalite ve Boyut Arttırmak----------------------------
    //----------------------------------------------------------------------------------------------

    public static string ResimDirekYuklemeResimKaliteBoyutArttirma(HttpPostedFile GozatYolu, string KlasorAdi, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png" | uzanti == ".swf" | uzanti == ".rar" | uzanti == ".zip" | uzanti == ".pdf" | uzanti == ".doc" | uzanti == ".docx" | uzanti == ".xls" | uzanti == ".xlsx" | uzanti == ".pps" | uzanti == ".ppsx" | uzanti == ".ppt" | uzanti == ".pptx" | uzanti == ".ico") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi; string random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------

            string tamYol = HttpContext.Current.Server.MapPath(KlasorAdi + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //---------------------------------------------------------------------------------------------------
            if (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png")
            {
                //-----------------------------------------------------------------------------------------------------------------------------------
                System.Drawing.Image imgPhotoVert1 = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
                System.Drawing.Image imgPhoto = null;
                imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert1, imgPhotoVert1.Width);
                imgPhotoVert1.Dispose();
                //-----------------------------------------------------------------------------------------------------------------------------------
                imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi, System.Drawing.Imaging.ImageFormat.Png);
                //-----------------------------------------------------------------------------------------------------------------------------------
                imgPhoto.Dispose();
                //-----------------------------------------------------------------------------------------------------------------------------------
            }
            //-----------------------------------------------------------------------------------------------------------------------------------
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {

                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimGenislikTekliKesmeResimKaliteBoyutArttirma(HttpPostedFile GozatYolu, string KlasorAdi, int Genislik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //-----------------------------------------------------------------------------------------------------------------------------------
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            //-----------------------------------------------------------------------------------------------------------------------------------
            System.Drawing.Image imgPhotoVert = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto = null;
            if (imgPhotoVert.Width > Genislik)
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, Genislik); }
            else
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, imgPhotoVert.Width); }
            imgPhotoVert.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png);
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            new FileInfo(tamYol).Delete();
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    public static string ResimGenislikCiftliKesmeResimKaliteBoyutArttirma(HttpPostedFile GozatYolu, string KlasorAdi, int KucukResimGenislik, int BuyukResimGenislik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //-----------------------------------------------------------------------------------------------------------------------------------
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            //-----------------------------------------------------------------------------------------------------------------------------------

            System.Drawing.Image imgPhotoVert = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto = null;
            if (imgPhotoVert.Width > KucukResimGenislik)
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, KucukResimGenislik); }
            else
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, imgPhotoVert.Width); }
            imgPhotoVert.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "k_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png);
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Dispose();

            //-----------------------------------------------------------------------------------------------------------------------------------
            System.Drawing.Image imgPhotoVert1 = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + ResimAdi);
            System.Drawing.Image imgPhoto1 = null;
            if (imgPhotoVert1.Width > BuyukResimGenislik)
            { imgPhoto1 = ResimBoyutlandirGenislik(imgPhotoVert1, BuyukResimGenislik); }
            else
            { imgPhoto1 = ResimBoyutlandirGenislik(imgPhotoVert1, imgPhotoVert1.Width); }
            imgPhotoVert1.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto1.Save(HttpContext.Current.Server.MapPath(resimKlasor) + "b_" + ResimAdi, System.Drawing.Imaging.ImageFormat.Png);
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto1.Dispose();

            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.ResimLogoDurum == true && !String.IsNullOrEmpty(Ayarlar.ResimLogo))
            {
                Bitmap bmp = new Bitmap(HttpContext.Current.Server.MapPath(resimKlasor + "/b_" + ResimAdi));
                Bitmap logo = new Bitmap(HttpContext.Current.Server.MapPath(ResimKlasoru + Ayarlar.ResimLogo));
                Fonksiyon.ResimCiz(bmp, logo).Save(HttpContext.Current.Server.MapPath(resimKlasor + "/L_" + ResimAdi));
            }
            //-----------------------------------------------------------------------------------------------------------------------------------

            new FileInfo(tamYol).Delete();
            tamYol = null;
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    //----------------------------------------------------------------------------------------------
    //-----------------------------------Resim Kalite ve Boyut Arttırmak---------------------------
    //----------------------------------------------------------------------------------------------




    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------
    //---------------------------Eski Resim Yükleme -------------------------------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------

    public static System.Drawing.Image Resize(System.Drawing.Image img, int en, int boy)
    {
        Bitmap bmp = new Bitmap(en, boy);
        Graphics graphic = Graphics.FromImage((System.Drawing.Image)bmp);
        graphic.InterpolationMode = InterpolationMode.HighQualityBicubic;
        graphic.SmoothingMode = SmoothingMode.HighQuality;
        graphic.PixelOffsetMode = PixelOffsetMode.HighQuality;
        graphic.CompositingQuality = CompositingQuality.HighQuality;
        graphic.DrawImage(img, 0, 0, en, boy);
        graphic.Dispose();
        return (System.Drawing.Image)bmp;
    }
    //-----------------------------------------------------------------------

    public static int[] ResimOran(double Y, double G, int sabit)
    {
        double ResimYukseklik = Y + 1;
        double ResimUzunluk = G;
        double oran = 0;
        if (ResimUzunluk > ResimYukseklik && ResimUzunluk > sabit)
        {
            oran = ResimUzunluk / ResimYukseklik;
            ResimUzunluk = sabit;
            ResimYukseklik = sabit / oran;
        }
        else if (ResimYukseklik > ResimUzunluk && ResimYukseklik > sabit)
        {
            oran = ResimYukseklik / ResimUzunluk;
            ResimYukseklik = sabit;
            ResimUzunluk = sabit / oran;
        }
        int yukseklik = Convert.ToInt32(ResimYukseklik);
        int genislik = Convert.ToInt32(ResimUzunluk);
        int[] geriDon = { genislik, yukseklik };
        return geriDon;
    }
    //----------------------------------------------------------------------------------------------

    public static string EskiResimGenislikTekliKesme(HttpPostedFile GozatYolu, string KlasorAdi, int Genislik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //-----------------------------------------------------------------------------------------------------------------------------------
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            //-----------------------------------------------------------------------------------------------------------------------------------
            string sonyol = null;
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            using (Bitmap OriginalBM = new Bitmap(tamYol))
            {
                double ResimYukseklik = OriginalBM.Height + 1;
                double ResimUzunluk = OriginalBM.Width;
                sonyol = HttpContext.Current.Server.MapPath(resimKlasor + "b_" + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
                //-----------------------------------------------------------------------------------------------------------------------------------
                if (Ayarlar.BuyukResimKaliteDurum == false)
                {
                    if (uzanti == ".jpg" | uzanti == ".jpeg")
                    { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, Genislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, Genislik)[1]).Save(sonyol, ImageFormat.Jpeg); }
                    else if (uzanti == ".gif")
                    { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, Genislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, Genislik)[1]).Save(sonyol, ImageFormat.Gif); }
                    else if (uzanti == ".png")
                    { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, Genislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, Genislik)[1]).Save(sonyol, ImageFormat.Png); }
                    else
                    { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, Genislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, Genislik)[1]).Save(sonyol, ImageFormat.Jpeg); }
                }
                else
                { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, Genislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, Genislik)[1]).Save(sonyol, ImageFormat.Png); }
                //-----------------------------------------------------------------------------------------------------------------------------------
                OriginalBM.Dispose();
            }
            //-----------------------------------------------------------------------------------------------------------------------------------

            new FileInfo(tamYol).Delete();
            tamYol = null;
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //-----------------------------------------------------------------------

    public static string EskiResimGenislikCiftliKesme(HttpPostedFile GozatYolu, string KlasorAdi, int KucukResimGenislik, int BuyukResimGenislik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //-----------------------------------------------------------------------------------------------------------------------------------
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            //-----------------------------------------------------------------------------------------------------------------------------------
            string sonyol = null;
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            using (Bitmap OriginalBM = new Bitmap(tamYol))
            {
                double ResimYukseklik = OriginalBM.Height + 1;
                double ResimUzunluk = OriginalBM.Width;

                //-----------------------------------------------------------------------------------------------------------------------------------
                sonyol = HttpContext.Current.Server.MapPath(resimKlasor + "k_" + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
                if (Ayarlar.KucukResimKaliteDurum == false)
                {
                    if (uzanti == ".jpg" | uzanti == ".jpeg")
                    { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, KucukResimGenislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, KucukResimGenislik)[1]).Save(sonyol, ImageFormat.Jpeg); }
                    else if (uzanti == ".gif")
                    { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, KucukResimGenislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, KucukResimGenislik)[1]).Save(sonyol, ImageFormat.Gif); }
                    else if (uzanti == ".png")
                    { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, KucukResimGenislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, KucukResimGenislik)[1]).Save(sonyol, ImageFormat.Png); }
                    else
                    { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, KucukResimGenislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, KucukResimGenislik)[1]).Save(sonyol, ImageFormat.Jpeg); }
                }
                else
                { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, KucukResimGenislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, KucukResimGenislik)[1]).Save(sonyol, ImageFormat.Png); }

                //-----------------------------------------------------------------------------------------------------------------------------------


                //-----------------------------------------------------------------------------------------------------------------------------------
                sonyol = HttpContext.Current.Server.MapPath(resimKlasor + "b_" + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
                if (Ayarlar.BuyukResimKaliteDurum == false)
                {
                    if (uzanti == ".jpg" | uzanti == ".jpeg")
                    { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, BuyukResimGenislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, BuyukResimGenislik)[1]).Save(sonyol, ImageFormat.Jpeg); }
                    else if (uzanti == ".gif")
                    { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, BuyukResimGenislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, BuyukResimGenislik)[1]).Save(sonyol, ImageFormat.Gif); }
                    else if (uzanti == ".png")
                    { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, BuyukResimGenislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, BuyukResimGenislik)[1]).Save(sonyol, ImageFormat.Png); }
                    else
                    { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, BuyukResimGenislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, BuyukResimGenislik)[1]).Save(sonyol, ImageFormat.Jpeg); }
                }
                else
                { Resize(OriginalBM, ResimOran(ResimYukseklik, ResimUzunluk, BuyukResimGenislik)[0], ResimOran(ResimYukseklik, ResimUzunluk, BuyukResimGenislik)[1]).Save(sonyol, ImageFormat.Png); }
                //-----------------------------------------------------------------------------------------------------------------------------------

                OriginalBM.Dispose();
            }
            //-----------------------------------------------------------------------------------------------------------------------------------

            if (Ayarlar.ResimLogoDurum == true && !String.IsNullOrEmpty(Ayarlar.ResimLogo))
            {
                Bitmap bmp = new Bitmap(HttpContext.Current.Server.MapPath(resimKlasor + "/b_" + ResimAdi));
                Bitmap logo = new Bitmap(HttpContext.Current.Server.MapPath(ResimKlasoru + Ayarlar.ResimLogo));
                Fonksiyon.ResimCiz(bmp, logo).Save(HttpContext.Current.Server.MapPath(resimKlasor + "/L_" + ResimAdi));
            }
            //-----------------------------------------------------------------------------------------------------------------------------------

            new FileInfo(tamYol).Delete();
            tamYol = null;
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //-----------------------------------------------------------------------

    public static string EskiResimOzelTekliKesme(HttpPostedFile GozatYolu, string KlasorAdi, int Genislik, int Yukseklik, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            int[] kucuk = new int[2]; kucuk[0] = Convert.ToInt32(Genislik); kucuk[1] = Convert.ToInt32(Yukseklik);
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //-----------------------------------------------------------------------------------------------------------------------------------
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            //-----------------------------------------------------------------------------------------------------------------------------------
            string sonyol = null;
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + "b_" + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            using (Bitmap OriginalBM = new Bitmap(tamYol))
            {
                double ResimYukseklik = OriginalBM.Height + 1;
                double ResimUzunluk = OriginalBM.Width;
                sonyol = HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
                if (Path.GetExtension(GozatYolu.FileName.ToLower()) != ".png") Resize(OriginalBM, kucuk[0], kucuk[1]).Save(sonyol, ImageFormat.Jpeg);
                else if (Path.GetExtension(GozatYolu.FileName.ToLower()) == ".png") Resize(OriginalBM, kucuk[0], kucuk[1]).Save(sonyol, ImageFormat.Jpeg);
                OriginalBM.Dispose();
            }
            new FileInfo(tamYol).Delete();
            tamYol = null;
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //-----------------------------------------------------------------------

    public static string EskiResimOzelCiftliKesme(HttpPostedFile GozatYolu, string KlasorAdi, int KucukGenislik, int KucukYukseklik, int BuyukGenislik, int BuyukYukseklik, string SiteResimAdi)
    {

        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".jpg" | uzanti == ".jpeg" | uzanti == ".gif" | uzanti == ".png") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            int[] kucuk = new int[2]; kucuk[0] = Convert.ToInt32(KucukGenislik); kucuk[1] = Convert.ToInt32(KucukYukseklik);
            int[] buyuk = new int[2]; buyuk[0] = Convert.ToInt32(BuyukGenislik); buyuk[1] = Convert.ToInt32(BuyukYukseklik);
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi, random = "";
            //-----------------------------------------------------------------------------------------------------------------------------------
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            //-----------------------------------------------------------------------------------------------------------------------------------
            string sonyol = null;
            string tamYol = System.Web.HttpContext.Current.Server.MapPath(resimKlasor + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
            GozatYolu.SaveAs(tamYol);
            using (Bitmap OriginalBM = new Bitmap(tamYol))
            {
                double ResimYukseklik = OriginalBM.Height + 1;
                double ResimUzunluk = OriginalBM.Width;
                sonyol = HttpContext.Current.Server.MapPath(resimKlasor + "k_" + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
                if (Path.GetExtension(GozatYolu.FileName.ToLower()) != ".png") Resize(OriginalBM, kucuk[0], kucuk[1]).Save(sonyol, ImageFormat.Jpeg);
                else if (Path.GetExtension(GozatYolu.FileName.ToLower()) == ".png") Resize(OriginalBM, kucuk[0], kucuk[1]).Save(sonyol, ImageFormat.Jpeg);

                sonyol = HttpContext.Current.Server.MapPath(resimKlasor + "b_" + random + Path.GetExtension(GozatYolu.FileName.ToLower()));
                if (Path.GetExtension(GozatYolu.FileName.ToLower()) != ".png") Resize(OriginalBM, buyuk[0], buyuk[1]).Save(sonyol, ImageFormat.Jpeg);
                else if (Path.GetExtension(GozatYolu.FileName.ToLower()) == ".png") Resize(OriginalBM, buyuk[0], buyuk[1]).Save(sonyol, ImageFormat.Jpeg);
                OriginalBM.Dispose();
            }
            //-----------------------------------------------------------------------------------------------------------------------------------

            if (Ayarlar.ResimLogoDurum == true && !String.IsNullOrEmpty(Ayarlar.ResimLogo))
            {
                Bitmap bmp = new Bitmap(HttpContext.Current.Server.MapPath(resimKlasor + "/b_" + ResimAdi));
                Bitmap logo = new Bitmap(HttpContext.Current.Server.MapPath(ResimKlasoru + Ayarlar.ResimLogo));
                Fonksiyon.ResimCiz(bmp, logo).Save(HttpContext.Current.Server.MapPath(resimKlasor + "/L_" + ResimAdi));
            }
            //-----------------------------------------------------------------------------------------------------------------------------------

            new FileInfo(tamYol).Delete();
            tamYol = null;
            return random + Path.GetExtension(GozatYolu.FileName.ToLower());
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    public static string CropResimGenislikKesme(string Tur, string KlasorAdi, int Genislik, string SiteResimAdi)
    {
        string ResimAdi = SiteResimAdi;
        string uzanti = ResimUzantisiBul(ResimAdi);
        if (SiteResimAdi.Length > 0 && (SiteResimAdi.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(SiteResimAdi.ToLower()) == 0))
        {
            //-----------------------------------------------------------------------------------------------------------------------------------
            string resimKlasor = KlasorAdi;
            //-----------------------------------------------------------------------------------------------------------------------------------
            System.Drawing.Image imgPhotoVert = System.Drawing.Image.FromFile(HttpContext.Current.Server.MapPath(resimKlasor) + Tur + ResimAdi);
            System.Drawing.Image imgPhoto = null;
            if (imgPhotoVert.Width > Genislik)
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, Genislik); }
            else
            { imgPhoto = ResimBoyutlandirGenislik(imgPhotoVert, imgPhotoVert.Width); }
            imgPhotoVert.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------
            if (uzanti == ".jpg" | uzanti == ".jpeg")
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + Tur + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            else if (uzanti == ".gif")
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + Tur + ResimAdi, System.Drawing.Imaging.ImageFormat.Gif); }
            else if (uzanti == ".png")
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + Tur + ResimAdi, System.Drawing.Imaging.ImageFormat.Png); }
            else
            { imgPhoto.Save(HttpContext.Current.Server.MapPath(resimKlasor) + Tur + ResimAdi, System.Drawing.Imaging.ImageFormat.Jpeg); }
            //-----------------------------------------------------------------------------------------------------------------------------------
            imgPhoto.Dispose();
            //-----------------------------------------------------------------------------------------------------------------------------------

            //-----------------------------------------------------------------------------------------------------------------------------------
            if (Ayarlar.ResimLogoDurum == true && !String.IsNullOrEmpty(Ayarlar.ResimLogo))
            {
                Bitmap bmp = new Bitmap(HttpContext.Current.Server.MapPath(resimKlasor + "/b_" + ResimAdi));
                Bitmap logo = new Bitmap(HttpContext.Current.Server.MapPath("/Uploads/GenelResim/" + Ayarlar.ResimLogo));
                Fonksiyon.ResimCiz(bmp, logo).Save(HttpContext.Current.Server.MapPath(resimKlasor + "/L_" + ResimAdi));
            }
            //-----------------------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (SiteResimAdi.ToLower().ToString() != "")
            {
                if (Fonksiyon.UploadSqlTemizle(SiteResimAdi.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(SiteResimAdi.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", SiteResimAdi.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------

    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------
    //--------------------------- Video Yükleme ----------------------------------------------------
    //----------------------------------------------------------------------------------------------
    //----------------------------------------------------------------------------------------------

    public static string VideoYukleme(HttpPostedFile GozatYolu, string KlasorAdi, string SiteResimAdi)
    {
        string ResimAdi = "";
        string uzanti = System.IO.Path.GetExtension(GozatYolu.FileName.ToLower());
        if (GozatYolu.ContentLength > 0 && (uzanti == ".aaf" | uzanti == ".3gp" | uzanti == ".asf" | uzanti == ".avchd" | uzanti == ".avi" | uzanti == ".cam" | uzanti == ".dat" | uzanti == ".mpeg" | uzanti == ".dsh" | uzanti == ".flv" | uzanti == ".m1v" | uzanti == ".m2v" | uzanti == ".swf" | uzanti == ".flr" | uzanti == ".mkv" | uzanti == ".wrap" | uzanti == ".mng" | uzanti == ".mov" | uzanti == ".mpeg" | uzanti == ".mp4" | uzanti == ".mxf" | uzanti == ".nsv" | uzanti == ".ogm" | uzanti == ".mng" | uzanti == ".rm" | uzanti == ".smi" | uzanti == ".wmv" | uzanti == ".divx" | uzanti == ".xvid") && (GozatYolu.FileName.ToLower().IndexOf(";") == -1) && (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) == 0))
        {
            string inputPath = ""; string outputPath = "";
            //---------------------------------------------------------------------------------------------------------------------
            string random = "";
            //-----------------------------------------------------------------------------------------------------------------------------------
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            try
            {
                if (Ayarlar.ResimBaslikAdiDurum == true && !String.IsNullOrEmpty(SiteResimAdi))
                { random = ResimSeoReplace(SiteResimAdi); }
                else if (Ayarlar.ResimYuklemeTur == true)
                { random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi; }
                else
                {
                    //random = Random(); 
                    random = ResimSeoReplace(GozatYolu.FileName.Replace(uzanti, "-")) + ResimRandomSayi;
                }
            }
            catch { random = Random(); }
            //--------------------------Resim Adı---------------------------------------------------------------------------------------------------------
            //-----------------------------------------------------------------------------------------------------------------------------------
            ResimAdi = random + Path.GetExtension(GozatYolu.FileName.ToLower());
            //---------------------------------------------------------------------------------------------------------------------
            string RootPath = HttpContext.Current.Server.MapPath(HttpContext.Current.Request.ApplicationPath);
            string filename = ResimAdi;
            //---------------------------------------------------------------------------------------------------------------------
            GozatYolu.SaveAs(HttpContext.Current.Server.MapPath("\\uploads\\GenelVideo\\") + "" + ResimAdi);
            string AppPath = HttpContext.Current.Request.PhysicalApplicationPath;
            inputPath = AppPath + "\\uploads\\GenelVideo\\" + "" + ResimAdi + "";
            outputPath = AppPath + "\\uploads\\GenelVideo\\" + "" + ResimAdi.Replace(uzanti, "") + ".flv";
            string fileargs = "";
            fileargs = " -i \"" + inputPath + "\" \"" + outputPath + "\"";
            fileargs = (" -i \"" + inputPath + "\" -ar 11025  -f flv -y \"") + outputPath + "\"";
            System.Diagnostics.Process proc = new System.Diagnostics.Process();
            proc.StartInfo.FileName = AppPath + "ffmpeg\\ffmpeg.exe";
            proc.StartInfo.Arguments = fileargs;
            proc.StartInfo.UseShellExecute = false;
            proc.StartInfo.CreateNoWindow = false;
            proc.StartInfo.RedirectStandardOutput = false;
            proc.Start();
            //---------------------------------------------------------------------------------------------------------------------
            System.Diagnostics.Process proc1 = new System.Diagnostics.Process();
            proc1.StartInfo.FileName = AppPath + "ffmpeg\\ffmpeg.exe";
            proc1.StartInfo.UseShellExecute = false;
            proc1.StartInfo.CreateNoWindow = false;
            proc1.StartInfo.RedirectStandardOutput = false;
            proc1.Start();
            //---------------------------------------------------------------------------------------------------------------------
        }
        else
        {
            if (GozatYolu.FileName.ToLower().ToString() != "")
            {

                if (Fonksiyon.UploadSqlTemizle(GozatYolu.FileName.ToLower()) != 0)
                {
                    //-----------------------------------------------//
                    PanelAyarlar.LogGenelRaporlari(Convert.ToInt32(HttpContext.Current.Session["YFZUserIDV"]), HttpContext.Current.Session["PKullaniciAdi"].ToString(), Fonksiyon.SqlTemizle(GozatYolu.FileName.ToLower()), 5);
                    //-----------------------------------------------//
                    Ayarlar.MailGonderme(Ayarlar.GenelMailAdresi.ToString(), "Zararlı Dosya Gönderimi", PanelAyarlar.UploadMailGonderme("Dosya", GozatYolu.FileName.ToLower()));
                    //-----------------------------------------------//
                }
                PanelAyarlar.IslemSonuMesaj("RH", HttpContext.Current.Request.ServerVariables["URL"] + "?" + PanelAyarlar.HataMesajReplace(HttpContext.Current.Request.ServerVariables["QUERY_STRING"]) + "");
                HttpContext.Current.Response.End();
            }
        }
        if (!String.IsNullOrEmpty(ResimAdi))
        { ResimAdi = ResimAdi.Replace(uzanti, ".flv"); }
        else
        { ResimAdi = ResimAdi; }
        return ResimAdi;
    }
    //----------------------------------------------------------------------------------------------
}