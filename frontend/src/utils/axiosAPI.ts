import axios from "axios";

export const getBaseUrl = () => {
    return import.meta.env.DEV ? "http://localhost:3000" : "/";
};

export const axiosAPI = axios.create({
    baseURL: getBaseUrl(),
});
