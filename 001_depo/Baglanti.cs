using System;
using System.Collections;
using System.Data;
using System.Data.SqlClient;
using System.Collections.Generic;
using System.Web;
using System.Web.UI;
using System.Web.UI.WebControls;

public class Baglanti
{   
    public static DataTable GetDataTable(string strsql)
    {
        SqlCommand cmd = new SqlCommand();
        SqlConnection conn = new SqlConnection();
        DataTable dt = new DataTable();
        SqlDataAdapter da = new SqlDataAdapter();
        da.SelectCommand = cmd;
        try
        {
            conn.ConnectionString = System.Configuration.ConfigurationManager.ConnectionStrings["ConnStr"].ToString();
            cmd.CommandText = strsql;
            cmd.CommandType = CommandType.Text;
            cmd.Connection = conn;
            da.Fill(dt);
            cmd.Dispose();
        }
        catch (Exception ex) { }
        finally
        {
            da.Dispose();
            cmd.Dispose();
            conn.Close();
            conn.Dispose();
        }
        return dt;
    }
    //-----------------------------------------------------------------------------

    public static string GetDataCell(string strsql)
    {
        DataTable table = GetDataTable(strsql);
        if (table.Rows.Count == 0)
        { return null; }
        return table.Rows[0][0].ToString();
    }
    //-----------------------------------------------------------------------------

    public static DataRow GetDataRow(string strsql)
    {
        SqlConnection conn = new SqlConnection();
        SqlCommand cmd = new SqlCommand();
        DataRow dr = null;
        DataTable dt = new DataTable();
        SqlDataAdapter da = new SqlDataAdapter();
        da.SelectCommand = cmd;
        try
        {
            conn.ConnectionString = System.Configuration.ConfigurationManager.ConnectionStrings["ConnStr"].ToString();
            cmd.CommandText = strsql;
            cmd.CommandType = CommandType.Text;
            cmd.Connection = conn;
            da.Fill(dt);
        }
        catch (Exception ex) { }
        finally
        {
            cmd.Dispose();
            da.Dispose();
            conn.Close();
            conn.Dispose();
        }
        if (dt.Rows.Count > 0)
        {  dr = dt.Rows[0]; }
        return dr;
    }
    //-----------------------------------------------------------------------------

    public static object ExecuteScaler(string strsql)
    {
        SqlConnection conn = new SqlConnection();
        SqlCommand cmd = new SqlCommand();
        object result = null;
        try
        {
            conn.ConnectionString = System.Configuration.ConfigurationManager.ConnectionStrings["ConnStr"].ToString();
            conn.Open();
            cmd.CommandText = strsql;
            cmd.CommandType = CommandType.Text;
            cmd.Connection = conn;
            result = cmd.ExecuteScalar();
        }
        catch (Exception ex) { }
        finally
        {
            cmd.Dispose();
            conn.Close();
            conn.Dispose();
        }
        return result;
    }
    //---------------------------------------------------------

    public static void ExecuteQuery(string strsql)
    {
        SqlConnection conn = new SqlConnection();
        SqlCommand cmd = new SqlCommand();
        try
        {
            conn.ConnectionString = System.Configuration.ConfigurationManager.ConnectionStrings["ConnStr"].ToString();
            conn.Open();
            cmd.CommandText = strsql;
            cmd.CommandType = CommandType.Text;
            cmd.Connection = conn;
            cmd.ExecuteNonQuery();
        }
        catch (Exception ex) { }
        finally
        {
            cmd.Dispose();
            conn.Close();
            conn.Dispose();
        }
    }
    //-----------------------------------------------------------------------------

    public static DataTable GetDataTableSayflama(string strsql, int page, int sayfadakiSayi)
    {
        SqlCommand cmd = new SqlCommand();
        SqlConnection conn = new SqlConnection();
        DataTable dt = new DataTable();
        SqlDataAdapter da = new SqlDataAdapter();
        da.SelectCommand = cmd;
        try
        {
            if (Convert.ToInt32(page) == 0)
            { page = 1; }
            conn.ConnectionString = System.Configuration.ConfigurationManager.ConnectionStrings["ConnStr"].ToString();
            cmd.CommandText = strsql;
            cmd.CommandType = CommandType.Text;
            cmd.Connection = conn;
            da.Fill((page - 1) * (sayfadakiSayi), sayfadakiSayi, dt);
            cmd.Dispose();
        }
        catch { }
        finally
        {
            da.Dispose();
            cmd.Dispose();
            conn.Close();
            conn.Dispose();
        }
        return dt;
    }
    //-----------------------------------------------------------------------------

}