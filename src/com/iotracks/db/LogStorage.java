package com.iotracks.db;

import com.iotracks.util.LogMessage;

import java.sql.*;
import java.util.ArrayList;
import java.util.List;
import java.util.StringJoiner;

public class LogStorage {

    private final static String DB_NAME = "iofog.db";
    private final static String TABLE_NAME = "logs";
    private final static String CONTAINER_ID_COLUMN_NAME = "container_id";
    private final static String LOG_LEVEL_COLUMN_NAME = "log_level";
    private final static String LOG_MESSAGE_COLUMN_NAME = "log_message";
    private final static String TIMESTAMP_COLUMN_NAME = "timestamp";

    private final static String SQL_DROP_TABLE = "drop table if exists " + TABLE_NAME;
    private final static String SQL_CREATE_TABLE = "create table " + TABLE_NAME + " (" + CONTAINER_ID_COLUMN_NAME + " string, " + LOG_LEVEL_COLUMN_NAME + " string, " + LOG_MESSAGE_COLUMN_NAME + " string, " + TIMESTAMP_COLUMN_NAME + " bigint) ";
    private final static String SQL_INSERT_TABLE = "insert into " + TABLE_NAME + " values ( \"%s\", \"%s\", \"%s\", %d )";
    private final static String SQL_SELECT_TABLE = "select * from " + TABLE_NAME + " where " + CONTAINER_ID_COLUMN_NAME + " in (%s) and " + TIMESTAMP_COLUMN_NAME + " >= %d and " + TIMESTAMP_COLUMN_NAME + " <= %d ORDER BY " + TIMESTAMP_COLUMN_NAME + " ASC";
    private final static String SQL_DELETE_TABLE = "delete from " + TABLE_NAME + " ";

    public static void create() {
        updateQuery(SQL_DROP_TABLE, SQL_CREATE_TABLE);
    }

    public static void addLog(LogMessage message) {
        System.out.println("LogStorage.addLog");
        updateQuery(String.format(SQL_INSERT_TABLE, message.getPublisherId(), message.getLevel(), message.getMessage(), message
                .getTimestamp()));
    }

    public static void deleteMessages(List<String> publishers, long timeframe) {
        System.out.println("LogStorage.deleteMessages");
        String sql = SQL_DELETE_TABLE;
        boolean and = false;
        if (publishers != null && !publishers.isEmpty()) {
            sql += " where ";
            sql += CONTAINER_ID_COLUMN_NAME + "in (" + formatList(publishers) + ")";
            and = true;
        }
        if (timeframe > 0) {
            if (and) {
                sql += " and ";
            } else {
                sql += " where ";
            }
            sql += TIMESTAMP_COLUMN_NAME + " <= " + timeframe;
        }
        System.out.println("LogStorage.deleteMessages SQL: " + sql);
        updateQuery(sql);
    }

    public static List<LogMessage> getMessages(List<String> publishers, long timeframestart, long timeframeend) {
        System.out.println("LogStorage.getMessages");
        try {
            Class.forName("org.sqlite.JDBC");
        } catch (ClassNotFoundException e) {
            System.err.println("ClassNotFoundException: " + e.getMessage());
        }

        Connection connection = null;
        try {
            connection = DriverManager.getConnection("jdbc:sqlite:" + DB_NAME);
            Statement statement = connection.createStatement();
            ResultSet rs = statement.executeQuery(String.format(SQL_SELECT_TABLE, formatList(publishers), timeframestart, timeframeend));
            List<LogMessage> messages = new ArrayList<>();
            while (rs.next()) {
                messages.add(new LogMessage(rs.getString(CONTAINER_ID_COLUMN_NAME), rs.getString(LOG_LEVEL_COLUMN_NAME), rs
                        .getString(LOG_MESSAGE_COLUMN_NAME), rs.getLong(TIMESTAMP_COLUMN_NAME)));
            }
            return messages;
        } catch (SQLException e) {
            // if the error message is "out of memory", it probably means no database file is found
            System.err.println(e.getMessage());
        } finally {
            try {
                if (connection != null) {
                    connection.close();
                }
            } catch (SQLException e) {
                System.err.println(e.getMessage());
            }
        }
        return null;
    }

    private static void updateQuery(String... sqlQueries) {
        try {
            Class.forName("org.sqlite.JDBC");
        } catch (ClassNotFoundException e) {
            System.err.println("ClassNotFoundException: " + e.getMessage());
        }

        Connection connection = null;
        try {
            connection = DriverManager.getConnection("jdbc:sqlite:" + DB_NAME);
            Statement statement = connection.createStatement();
            for (String sqlQuery : sqlQueries) {
                statement.executeUpdate(sqlQuery);
            }
        } catch (SQLException e) {
            System.err.println(e.getMessage());
        } finally {
            try {
                if (connection != null) {
                    connection.close();
                }
            } catch (SQLException e) {
                System.err.println(e.getMessage());
            }
        }
    }

    private static String formatList(List<String> list) {
        StringJoiner sj = new StringJoiner(",");
        list.forEach(listItem -> sj.add(listItem));
        return sj.toString();
    }
}
