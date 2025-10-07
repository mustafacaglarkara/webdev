using System;
using System.Data;
using System.Collections.Generic;
using System.Linq;
using System.Web;
using System.Drawing;
using System.Web.UI.WebControls;

public class HataMail
{
    private static DataRow rs = null;
    //------------------------------------------------------------------------

    public static void BaglantiTablosu()
    { rs = Baglanti.GetDataRow("select * from Tbl_HataMailAyarlari order by id asc"); }
    //------------------------------------------------------------------------

    public static Boolean MailEnable
    { get { BaglantiTablosu(); return  Boolean.Parse(rs["MailEnable"].ToString()); } }
    //-----------------------------------------------------------------------

    public static string MailTime
    { get { BaglantiTablosu(); return rs["MailTime"].ToString(); } }
    //-----------------------------------------------------------------------

    public static string MailAdress
    { get { BaglantiTablosu(); return rs["MailAdress"].ToString(); } }
    //-----------------------------------------------------------------------

    public static string MailEmail
    { get { BaglantiTablosu(); return rs["MailEmail"].ToString(); } }
    //-----------------------------------------------------------------------

    public static void HataMailGonder(string exMessage, string exSubject)
    {
        BaglantiTablosu();
        if (Boolean.Parse(rs["MailEnable"].ToString())==true)
        {
            string icerik = rs["MailMessage"].ToString().Replace("{mailcompetent}", rs["MailCompetent"].ToString()).Replace("{hatakonu}", exSubject).Replace("{hatamesajı}", exMessage).Replace("{hatadomain}", "http://" + HttpContext.Current.Request.ServerVariables["HTTP_HOST"]);
            icerik = icerik.Replace("{hatasayfa}", "http://" + HttpContext.Current.Request.ServerVariables["HTTP_HOST"] + HttpContext.Current.Request.RawUrl);
            icerik = icerik.Replace("{mailtime}", rs["MailTime"].ToString()).Replace("{mailadress}", rs["MailAdress"].ToString());
            Ayarlar.MailGonderme(rs["mailAdress"].ToString(), "Hata Raporu", icerik);
        }                                                              
    }


}